

let errorDisplay = document.getElementById("errorMessage");


//  handling slider betweeen reistration and login form starts here
const switchElement = document.querySelector(".switch");
const sliderButtons = document.querySelectorAll(".option");



sliderButtons[0].addEventListener("click", () => {
    switchElement.classList.remove("register-active");
    document.querySelector(".loginDiv").style.display = "none";
    document.querySelector(".registerDiv").style.display = "block";
    document.querySelector(".registerDiv").classList.add("fadeInReg");
});

sliderButtons[1].addEventListener("click", () => {
    switchElement.classList.add("register-active");
    document.querySelector(".registerDiv").style.display = "none";
    document.querySelector(".loginDiv").style.display = "block";
    document.querySelector(".loginDiv").classList.add("fadeInLog");
});


//  handling slider betweeen reistration and login form ENDS here


/*   CHECKING IF USERNAME ALREADY EXIST IN THE DATABASE*/

let timeout

const usernameInput = document.getElementById("userName");
usernameInput.addEventListener("input", (e) => {

    clearTimeout(timeout)

    const username = e.target.value.trim();

    if (username.length < 3) {
        return;
    }

    timeout = setTimeout(() => {

        fetch("/userName", {

            method: "POST",
            headers: {
                "Content-Type": "text/plain"
            },
            body: username
        }).then((res) => {
            return res.text().then((data) => ({
                status: res.status,
                message: data
            }));
        }).then((result) => {
            if (result.status === 409 || result.status === 400) {
                usernameInput.style.border = "2px solid red";
                errorDisplay.innerText = result.message;
                errorDisplay.style.display = "block"
            } else if (result.status === 200) {
                usernameInput.style.border = "";
                errorDisplay.style.display = "none"
            }
        })

    }, 400)
})


//  sending user registration details to backend STARTS here

const form = document.querySelector(".register");


form.addEventListener("submit", (e) => {
    e.preventDefault();

    const userData = new FormData(form);


    const formObject = Object.fromEntries(userData);



    fetch("/register", {
        method: "POST",
        headers: {
            "Content-Type": "application/json"
        },
        body: JSON.stringify(formObject)

    }).then((res) => {
        return res.json().then((data) => ({
            status: res.status,
            res: data
        }));

    }).then((result) => {

        clearRegistrationErrors();

        if (result.status === 409 || result.status === 400) {
            ele = document.getElementById(result.res.ElementId);

            ele.style.border = "2px solid red";
            errorDisplay.innerText = result.res.Message;
            errorDisplay.style.display = "block"
        } else if (result.status === 200) {
            errorDisplay.style.display = "none"
            window.location.href = "/";
        } else {
            errorDisplay.innerText = "Something went wrong. Please try again.";
            errorDisplay.style.display = "block";
        }

        console.log(ele)
        console.log(result.status)
        console.log(result.res.Message)
        console.log(result.res.ElementId)
    });

});

function clearRegistrationErrors() {
    document.querySelectorAll(".register input").forEach((input) => {
        input.style.border = "";
    });

    errorDisplay.style.display = "none";
    errorDisplay.innerText = "";
}


//  sending user registration details to backend ENDS here