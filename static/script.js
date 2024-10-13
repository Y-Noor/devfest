const videoElement = document.getElementById("videoElement");

navigator.mediaDevices.getUserMedia({ video: true })
    .then(stream => {
        videoElement.srcObject = stream;
    })
    .catch(error => {
        console.error("Error accessing camera:", error);
    });