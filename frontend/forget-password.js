



let ResetPasword = document.getElementById("forgetPassword");

ResetPasword.addEventListener("submit", (e) => {
    e.preventDefault();


    let data = new FormData(ResetPasword)
    fetch("/forget-password", {
        method: "POST",
        body: data
    }).then((res) => {
        return res.text().then((data) => ({
            status: res.status,
            res: data
        }));
    }).then((result) => {
        if (result.status == 200) {
            document.getElementById("email").textContent = result.res
            document.querySelectorAll(".close")[0].style.display = "none"
            document.querySelectorAll(".close")[1].style.display = "none"
            document.querySelector(".open").style.display = "block"


        }
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