package backEnd

import "strings"


//

func MaskEmail(email string) string {
	parts := strings.SplitN(email, "@", 2)

	if len(parts) != 2 {
		return ""
	}

	local := parts[0]
	domain := parts[1]

	if len(local) <= 2 {
		return local[:1] + "***@" + domain
	}

	return local[:1] + "********" + local[len(local)-2:] + "@" + domain
}
