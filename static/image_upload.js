const selectedFiles = new Map();
const maxTotalSize = 20 * 1024 * 1024; // 20 MB
let totalSize = 0;

function updateFileList() {
  const input = document.getElementById("files");
  const warning = document.getElementById("warning");
  const submitButton = document.getElementById("submitButton");
  const previewContainer = document.getElementById("previewContainer");
  previewContainer.style.display = "block"
  warning.style.display = "block"

  for (const file of input.files) {
    if (selectedFiles.has(file.name)) continue;

    selectedFiles.set(file.name, file.size);
    totalSize += file.size;

    const reader = new FileReader();
    reader.onload = (e) => {
      const previewDiv = document.createElement("div");

      const imageBox = document.createElement("div");
      imageBox.classList.add("image-box");

      const img = document.createElement("img");
      img.src = e.target.result;
      img.classList.add("image-preview");
      imageBox.appendChild(img);

      const infoBox = document.createElement("div");
      infoBox.classList.add("image-info-box");

      const fileName = truncateFileName(file.name);
      const title = document.createElement("div");
      title.textContent = fileName;
      title.classList.add("image-info-title");

      const size = document.createElement("div");
      size.textContent = formatSize(file.size);
      size.classList.add("image-info-size");

      const deleteButton = document.createElement("button");
      deleteButton.textContent = "×";
      deleteButton.classList.add("delete-button");
      deleteButton.onclick = () => deleteFile(file.name, previewDiv);

      infoBox.appendChild(title);
      infoBox.appendChild(size);
      infoBox.appendChild(deleteButton);

      previewDiv.appendChild(imageBox);
      previewDiv.appendChild(infoBox);

      previewContainer.appendChild(previewDiv);
    };
    reader.readAsDataURL(file);
  }

  input.value = "";

  if (totalSize > maxTotalSize) {
    showElement("warning");
    warning.textContent = "Total file size exceeds 20 MB. Please remove some files.";
    submitButton.disabled = true;
  } else {
    warning.textContent = "";
    submitButton.disabled = false;
  }
}

function deleteFile(fileName, previewDiv) {
  if (selectedFiles.has(fileName)) {
    totalSize -= selectedFiles.get(fileName);
    selectedFiles.delete(fileName);
  }
  previewDiv.remove();

  const warning = document.getElementById("warning");
  const submitButton = document.getElementById("submitButton");

  if (totalSize > maxTotalSize) {
    showElement("warning");
    warning.textContent = "Total file size exceeds 20 MB. Please delete some files.";
    submitButton.disabled = true;
  } else {
    warning.textContent = "";
    submitButton.disabled = false;
  }
}

function formatSize(bytes) {
  const units = ["B", "KB", "MB"];
  let unitIndex = 0;

  while (bytes >= 1024 && unitIndex < units.length - 1) {
    bytes /= 1024;
    unitIndex++;
  }

  return `${bytes.toFixed(2)} ${units[unitIndex]}`;
}

function truncateFileName(name) {
  const parts = name.split(".");
  const extension = parts.pop();
  const baseName = parts.join(".");
  return baseName.length > 10 ? `${baseName.substring(0, 10)}...${extension}` : name;
}