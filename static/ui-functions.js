/* DARK MODE */
// Makes the script run only after html content is loaded
document.addEventListener('DOMContentLoaded', function () {
  const theme = localStorage.getItem('theme');
  const icon = document.getElementById('themeIcon');
  if (theme === 'dark') {
    document.body.classList.add('dark-mode');
    icon.textContent = 'light_mode';
  } else {
    icon.textContent = 'dark_mode';
  }
});

// Change between light and dark mode
function toggleDarkMode() {
  const body = document.body;
  const icon = document.getElementById('themeIcon');
  body.classList.toggle('dark-mode');
  const currentTheme = body.classList.contains('dark-mode') ? 'dark' : 'light';
  localStorage.setItem('theme', currentTheme);
  icon.textContent = currentTheme === 'dark' ? 'light_mode' : 'dark_mode';
}


/* NEW THREAD MODAL */
// Get the modal
var modal = document.getElementById("newpostModal");

// Get the button that opens the modal
var btn = document.getElementById("modalBtn");

// Get the <span> element that closes the modal
var span = document.getElementsByClassName("close")[0];

// When the user clicks the button, open the modal 
btn.onclick = function () {
  modal.style.display = "block";
}

// When the user clicks on <span> (x), close the modal
span.onclick = function () {
  modal.style.display = "none";
}

// When the user clicks anywhere outside of the modal, close it
window.onclick = function (event) {
  if (event.target == modal) {
    modal.style.display = "none";
  }
}

/* REPLY */
// Open and close reply-to-reply form with "Reply" button
document.addEventListener("DOMContentLoaded", function () {
  const buttons = document.querySelectorAll(".reply-button");
  buttons.forEach(button => {
    button.addEventListener("click", function () {
      const formContainer = this.closest("table").nextElementSibling;
      if (formContainer.style.display === "none") {
        formContainer.style.display = "block";
      } else {
        formContainer.style.display = "none";
      }
    });
  });
});