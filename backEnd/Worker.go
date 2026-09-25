package backEnd

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

/*       CREATING WORKER TO SEND   VERIFICATION CODE TO USER EMAIL ACCOUNT STARTS HERE    */

func GetPendingEmails(ctx context.Context, pool *pgxpool.Pool) ([]Email, error) {

	tx, err := pool.Begin(ctx)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback(ctx)

	rows, err := tx.Query(ctx, `SELECT id, recipient_email, subject, body, attempts 
	FROM email_outbox 
	WHERE status = 'pending' 
	ORDER BY created_at 
	LIMIT 5 
	FOR UPDATE SKIP LOCKED
	`)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	emails, err := pgx.CollectRows(
		rows,
		pgx.RowToStructByName[Email],
	)

	if err != nil {
		return nil, err
	}

	for _, email := range emails {

		_, err := tx.Exec(ctx, `
            UPDATE email_outbox
            SET status = 'processing'
            WHERE id = $1
        `, email.ID)

		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return emails, nil
}

func ProcessPendingEmails(ctx context.Context, pool *pgxpool.Pool) {

	emails, err := GetPendingEmails(ctx, pool)

	if err != nil {
		fmt.Println("Error getting the pending emails: ", err)
		return
	}

	for _, email := range emails {
		err := SendMail(email)

		if err != nil {
			fmt.Println("failed to send email: ", email.ID)

			// update email_outbox for failed mail verication sending
			_, updateErr := pool.Exec(ctx, `UPDATE email_outbox 
			    SET attempts = attempts + 1,
				    status = 'pending',
				    last_error = $1
			    WHERE id = $2`,
				err.Error(),
				email.ID,
			)

			if updateErr != nil {
				fmt.Println("failed to update record error: ", updateErr)
			}
			continue
		}

		// update email_outbox for a successful  mail verication sending
		_, err = pool.Exec(ctx, `UPDATE email_outbox 
			    SET status = 'sent',
				    sent_at = NOW(),
					attempts = attempts + 1,
					last_error = NULL
			    WHERE id = $1`,
			email.ID,
		)

		if err != nil {
			fmt.Println("failed to email as sent: ", err)
		}
	}
}

func StartEmailWorker(ctx context.Context, pool *pgxpool.Pool) {

	ticker := time.NewTicker(5 * time.Second)

	defer ticker.Stop()

	for {

		select {

		case <-ticker.C:
			ProcessPendingEmails(ctx, pool)

		case <-ctx.Done():
			fmt.Println("EMAIL WORKER SHOTTING DOWN")
			return
		}
	}
}

/*       CREATING WORKER TO SEND   VERIFICATION CODE TO USER EMAIL ACCOUNT ENDS HERE    */
