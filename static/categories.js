function updateCategory() {
    const select = document.getElementById('categorySelector');
    const selectedCategory = select.value;
    const categoriesInput = document.getElementById('categories');
    if (!categoriesInput.value.includes(selectedCategory)) {
        categoriesInput.value += selectedCategory + ' ';
      }
    select.selectedIndex = 0;
  }