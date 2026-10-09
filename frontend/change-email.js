

let errorDisplay = document.getElementById("errorMessage");

let resendCode = document.getElementById("changeEmail");

resendCode.addEventListener("submit", (e) => {
    e.preventDefault();


    let data = new FormData(resendCode)
    fetch("/change-email", {
        method: "POST",
        body: data
    }).then((res) => {
        return res.text().then((data) => ({
            status: res.status,
            res: data
        }));
    }).then((result) => {

        if (result.status === 400 || result.status === 409) {
            errorMessage.style.visibility = "visible";
            errorMessage.textContent = result.res;

        } else if (result.status === 200) {
            errorMessage.style.visibility = "none";
            window.location.href = "/verify-email";

        } else if (result.status === 401 || result.status === 404) {
            errorMessage.style.visibility = "none";
            window.location.href = "/onboarding";

        } else {
            errorDisplay.textContent = "Something went wrong. Please try again";
            errorDisplay.style.visibility = "visible";
        }

    })
})