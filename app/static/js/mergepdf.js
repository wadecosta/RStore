const fileInput = document.getElementById("pdfs");
const pdfList = document.getElementById("pdf-list");
const mergeForm = document.getElementById("mergeForm");

let selectedFiles = [];

fileInput.addEventListener("change", () => {

    selectedFiles = Array.from(fileInput.files).map(file => ({
        id: crypto.randomUUID(),
        file: file
    }));

    renderFiles();
});

function renderFiles() {

    pdfList.innerHTML = "";

    selectedFiles.forEach((item, index) => {

        const col = document.createElement("div");

        col.className = "col-md-3";
        col.draggable = true;
        col.dataset.id = item.id;

        col.innerHTML = `
            <div class="card pdf-card h-100">

                <div class="card-body text-center">

                    <div class="pdf-preview">
                        📄
                    </div>

                    <h5>${index + 1}</h5>

                    <div class="text-break">
                        ${item.file.name}
                    </div>

                    <small class="text-secondary">
                        ${(item.file.size / 1024 / 1024).toFixed(2)} MB
                    </small>

                </div>

            </div>
        `;

        addDragEvents(col);

        pdfList.appendChild(col);
    });
}

function addDragEvents(element) {

    element.addEventListener("dragstart", () => {
        element.classList.add("dragging");
    });

    element.addEventListener("dragend", () => {
        element.classList.remove("dragging");
    });

    element.addEventListener("dragover", e => {
        e.preventDefault();
    });

    element.addEventListener("drop", e => {

        e.preventDefault();

        const dragging =
            document.querySelector(".dragging");

        const fromID =
            dragging.dataset.id;

        const toID =
            element.dataset.id;

        const fromIndex =
            selectedFiles.findIndex(f => f.id === fromID);

        const toIndex =
            selectedFiles.findIndex(f => f.id === toID);

        const moved =
            selectedFiles.splice(fromIndex, 1)[0];

        selectedFiles.splice(toIndex, 0, moved);

        renderFiles();
    });
}


mergeForm.addEventListener("submit", async (e) => {

    e.preventDefault();

    if (selectedFiles.length < 2) {
        alert("Please select at least 2 PDFs.");
        return;
    }

    const formData = new FormData();

    formData.append(
        "pdf_order",
        selectedFiles.map(f => f.id).join(",")
    );

    selectedFiles.forEach(item => {
        formData.append("pdfs", item.file);
        formData.append("pdf_ids", item.id);
    });

    const response = await fetch("/merge-pdf", {
        method: "POST",
        body: formData
    });

    if (!response.ok) {
        alert(await response.text());
        return;
    }

    const blob = await response.blob();

    const url =
        window.URL.createObjectURL(blob);

    const a =
        document.createElement("a");

    a.href = url;
    a.download = "merged.pdf";
    a.click();

    window.URL.revokeObjectURL(url);
});
