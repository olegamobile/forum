function updateCategory(fieldName, selectorName) {
    const select = document.getElementById(selectorName);
    const selectedCategory = select.value;
    const categoriesInput = document.getElementById(fieldName);
    if (!categoriesInput.value.includes(selectedCategory)) {
        categoriesInput.value += selectedCategory + ' ';
      }
    select.selectedIndex = 0;
  }