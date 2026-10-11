

let errorMessage = document.getElementById("errorMessage");
const resultElement = document.getElementById("result");

// Handling live count down STARTS here
const countdown = document.getElementById("countdown");

const expiresAt = new Date(
    countdown.dataset.expiresAt
);

function updateCountdown() {
    const now = new Date();

    const remaining = expiresAt - now;

    if (remaining <= 0) {
        countdown.textContent = "Expired";
        clearInterval(timer);
        return;
    }

    const totalSeconds = Math.floor(remaining / 1000);

    const minutes = Math.floor(totalSeconds / 60);
    const seconds = totalSeconds % 60;

    countdown.textContent =
        `${minutes}:${seconds.toString().padStart(2, "0")}`;
}


const timer = setInterval(updateCountdown, 1000);

updateCountdown();

// Handling live count down ENDS here

const token = new URLSearchParams(window.location.search).get("token");

let changePassword = document.getElementById("forgetPassword");
changePassword.addEventListener("submit", (e) => {
    e.preventDefault();


    let data = new FormData(changePassword)
    fetch(`/change-password?token=${encodeURIComponent(token)}`, {
        method: "POST",
        body: data
    }).then((res) => {
        return res.text().then((data) => ({
            status: res.status,
            res: data
        }));
    }).then((result) => {

        if (result.status === 400 || result.status === 404) {
            errorMessage.style.visibility = "visible";
            errorMessage.textContent = result.res;

        } else if (result.status === 200) {
            errorMessage.style.visibility = "hidden";

            resultElement.textContent = "Password successfully changed. ";
            resultElement.append(document.createElement("br"));

            const loginLink = document.createElement("a");
            loginLink.href = "/onboarding";
            loginLink.textContent = "Go to login";

            resultElement.append(loginLink);

            changePassword.reset();
            clearInterval(timer);
        } else {
            errorDisplay.textContent = "Something went wrong. Please try again";
            errorDisplay.style.visibility = "visible";
        }

    })
})