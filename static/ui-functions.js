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

/* REPLY */
// Open and close reply-to-reply form with "Reply" button
document.addEventListener("DOMContentLoaded", function () {
  const buttons = document.querySelectorAll(".reply-button");
  buttons.forEach(button => {
    button.addEventListener("click", function () {
      // Find the closest thread and then look for the reply-form-container sibling
      const formContainer = this.closest(".reply-and-form").querySelector(".reply-form-container");

      // Toggle the display of the form container
      if (formContainer.style.display === "none" || formContainer.style.display === "") {
        formContainer.style.display = "block";
      } else {
        formContainer.style.display = "none";
      }
    });
  });
});
