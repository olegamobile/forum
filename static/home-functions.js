/* NEW THREAD MODAL */
// Get the modal
var modal = document.getElementById("newpostModal");

// Get the button that opens the modal
var btn = document.getElementById("modalBtn");

// Get the <span> element that closes the modal
var span = document.getElementsByClassName("close")[0];

// When the user clicks the button, open the modal 
if (btn) {
  btn.onclick = function () {
    modal.style.display = "block";
  }
}

// When the user clicks on <span> (x), close the modal
if (span) {
  span.onclick = function () {
    modal.style.display = "none";
  }
}

// When the user clicks anywhere outside of the modal, close it
window.onclick = function (event) {
  if (event.target == modal) {
    modal.style.display = "none";
  }
}

// Toggle filter div visibility with filter button
document.addEventListener("DOMContentLoaded", function () {
  const filterButton = document.getElementById("filter-button");
  const filterDiv = document.getElementById("show-filter");

  if (filterDiv) {
    // Check localStorage for the visibility state
    const savedState = localStorage.getItem("filter-visible");
    if (savedState === "true") {
      filterDiv.classList.add("visible");
      updateButton(filterButton, "filter_alt_off", "Hide filter");
    } else {
      filterDiv.classList.remove("visible");
      updateButton(filterButton, "filter_alt", "Show filter");
    }
  }

  if (filterButton && filterDiv) {
    filterButton.addEventListener("click", function () {

      // Toggle the visibility class
      if (filterDiv.classList.contains("visible")) {
        filterDiv.classList.remove("visible");
        localStorage.setItem("filter-visible", "false");
        updateButton(filterButton, "filter_alt", "Show filter");
      } else {
        filterDiv.classList.add("visible");
        localStorage.setItem("filter-visible", "true");
        updateButton(filterButton, "filter_alt_off", "Hide filter");
      }
    });
  }

  // Function to update button content
  function updateButton(button, iconName, text) {
    const span = button.querySelector("span.material-symbols-outlined"); // Find the icon span
    if (span) span.textContent = iconName; // Update the icon
    filterButton.lastChild.textContent = ` ${text}`; // Update the text after the icon
  }
});