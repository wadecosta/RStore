const fileInput = document.getElementById("pdfs");
const pdfList = document.getElementById("pdf-list");
const mergeForm = document.getElementById("mergeForm");

let selectedFiles = [];

/*
 * Generate a simple unique ID.
 * Doesn't depend on crypto.randomUUID().
 */
function generateId() {
    return Date.now().toString(36) + Math.random().toString(36).substring(2);
}


/*
 * Handle PDF selection.
 */
fileInput.addEventListener("change", () => {

    const files = Array.from(fileInput.files);

    selectedFiles = files.map(file => ({
        id: generateId(),
        file: file
    }));

    renderFiles();
});


/*
 * Display selected PDFs.
 */
function renderFiles() {

    pdfList.innerHTML = "";

    if (selectedFiles.length === 0) {
        pdfList.innerHTML = `
            <div class="col-12">
                <div class="empty-state">
                    No PDFs selected.
                </div>
            </div>
        `;

        return;
    }

    selectedFiles.forEach((item, index) => {

        const col = document.createElement("div");

        col.className = "col-md-3";
        col.draggable = true;
        col.dataset.id = item.id;

        col.innerHTML = `
            <div class="card pdf-card h-100">

                <div class="card-body text-center">

                    <div class="pdf-preview">
                        PDF
                    </div>

                    <div class="pdf-number">
                        ${index + 1}
                    </div>

                    <div class="pdf-name text-break">
                        ${escapeHtml(item.file.name)}
                    </div>

                    <small class="text-secondary">
                        ${formatFileSize(item.file.size)}
                    </small>

                </div>

            </div>
        `;

        addDragEvents(col);

        pdfList.appendChild(col);
    });
}


/*
 * Prevent PDF filenames from being interpreted as HTML.
 */
function escapeHtml(value) {

    const div = document.createElement("div");

    div.textContent = value;

    return div.innerHTML;
}


/*
 * Format file size.
 */
function formatFileSize(bytes) {

    if (bytes === 0) {
        return "0 Bytes";
    }

    const units = [
        "Bytes",
        "KB",
        "MB",
        "GB"
    ];

    const index =
        Math.floor(Math.log(bytes) / Math.log(1024));

    return (
        (bytes / Math.pow(1024, index)).toFixed(2)
        + " "
        + units[index]
    );
}


/*
 * Drag and drop functionality.
 */
function addDragEvents(element) {

    element.addEventListener("dragstart", () => {

        element.classList.add("dragging");
    });


    element.addEventListener("dragend", () => {

        element.classList.remove("dragging");
    });


    element.addEventListener("dragover", event => {

        event.preventDefault();
    });


    element.addEventListener("drop", event => {

        event.preventDefault();

        const dragging =
            document.querySelector(".dragging");

        if (!dragging || dragging === element) {
            return;
        }

        const fromID =
            dragging.dataset.id;

        const toID =
            element.dataset.id;

        const fromIndex =
            selectedFiles.findIndex(
                file => file.id === fromID
            );

        const toIndex =
            selectedFiles.findIndex(
                file => file.id === toID
            );

        if (fromIndex === -1 || toIndex === -1) {
            return;
        }

        const moved =
            selectedFiles.splice(fromIndex, 1)[0];

        selectedFiles.splice(toIndex, 0, moved);

        renderFiles();
    });
}


/*
 * Submit PDFs to the Go backend.
 */
mergeForm.addEventListener("submit", async event => {

    event.preventDefault();

    if (selectedFiles.length < 2) {

        alert("Please select at least 2 PDFs.");

        return;
    }

    const formData = new FormData();

    /*
     * Send the order separately so the backend
     * knows which PDF should come first.
     */
    formData.append(
        "pdf_order",
        selectedFiles
            .map(file => file.id)
            .join(",")
    );


    /*
     * Send each PDF and its ID.
     */
    selectedFiles.forEach(item => {

        formData.append(
            "pdfs",
            item.file,
            item.file.name
        );

        formData.append(
            "pdf_ids",
            item.id
        );
    });


    try {

        const response = await fetch(
            "/merge-pdf",
            {
                method: "POST",
                body: formData
            }
        );


        if (!response.ok) {

            alert(await response.text());

            return;
        }


        const blob =
            await response.blob();

        const url =
            window.URL.createObjectURL(blob);

        const a =
            document.createElement("a");

        a.href = url;
        a.download = "merged.pdf";

        document.body.appendChild(a);

        a.click();

        a.remove();

        window.URL.revokeObjectURL(url);

    } catch (error) {

        console.error(error);

        alert(
            "An error occurred while merging the PDFs."
        );
    }
});


/*
 * Show the initial empty state.
 */
renderFiles();
