

let errorDisplay = document.getElementById("errorMessage");

console.log(errorDisplay)

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
        if (result.status == 200 || result.status == 404) {
            document.getElementById("email").textContent = result.res
            document.querySelectorAll(".close")[0].style.display = "none"
            document.querySelectorAll(".close")[1].style.display = "none"
            document.querySelector(".open").style.display = "block"
        }
        if (result.status == 400) {
            errorDisplay.style.visibility = "visible";
            errorDisplay.textContent = result.res;
            console.log(errorDisplay)

        } else {
            errorDisplay.textContent = "Something went wrong. Please try again";
            errorDisplay.style.visibility = "visible";
        }

    })
})