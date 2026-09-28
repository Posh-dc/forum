

let errorDisplay = document.getElementById("errorMessage");

// controling user input STARTS here 
const inputs = document.querySelectorAll(".inpt input");

inputs.forEach((input, index) => {

    input.addEventListener("input", () => {
        input.value = input.value.replace(/\D/g, "");

        if (input.value !== "" && index < inputs.length - 1) {
            inputs[index + 1].focus();
        }
    });

    input.addEventListener("keydown", (event) => {
        if (
            event.key === "Backspace" &&
            input.value === "" &&
            index > 0
        ) {
            inputs[index - 1].focus();
        }
    });
});

// controling user input ENDS here 


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

updateCountdown();

const timer = setInterval(updateCountdown, 1000);

// Handling live count down ENDS here


// Sending the verification to backend starts here

let form = document.getElementById("form")
form.addEventListener("submit", (e) => {
    e.preventDefault();



    const formData = new FormData(form);

    for (const [key, value] of formData) {
        console.log(key, value);
    }

    fetch("/verify-email", {
        method: "POST",
        body: formData
    }).then((res) => {
        return res.text().then((data) => ({
            status: res.status,
            res: data
        }));
    }).then((result) => {

        errorMessage.textContent = result.res;
        errorMessage.style.visibility = "visible";
        console.log(result.status)
        console.log(result.res)
    })
});




// Sending the verification to backend starts here