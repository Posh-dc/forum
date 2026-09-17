const form = document.querySelector(".register");

console.log(form)

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
        return res.text().then((message) => ({
            status: res.status,
            message: message
        }));

    }).then((data) => {
        console.log(data.status)
        console.log(data.message)
    });

});