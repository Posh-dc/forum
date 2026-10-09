let changePassword = document.getElementById("forgetPassword");



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

changePassword.addEventListener("submit", (e) => {
    e.preventDefault();


    let data = new FormData(changePassword)
    fetch("/change-password", {
        method: "POST",
        body: data
    }).then((res) => {
        return res.text().then((data) => ({
            status: res.status,
            res: data
        }));
    }).then((result) => {

        // if (result.status === 400 || result.status === 409) {
        //     errorMessage.style.visibility = "visible";
        //     errorMessage.textContent = result.res;

        // } else if (result.status === 200) {
        //     errorMessage.style.visibility = "none";
        //     window.location.href = "/verify-email";

        // } else if (result.status === 401 || result.status === 404) {
        //     errorMessage.style.visibility = "none";
        //     window.location.href = "/onboarding";

        // } else {
        //     errorDisplay.textContent = "Something went wrong. Please try again";
        //     errorDisplay.style.visibility = "visible";
        // }

    })
})