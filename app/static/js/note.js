document.addEventListener("DOMContentLoaded", () => {

	/* Edit */
	const editModal = document.getElementById("editNoteModal");

	editModal.addEventListener("show.bs.modal", function (event) {
		const button = event.relatedTarget;
		document.getElementById("edit-note-id").value = button.dataset.id;
		document.getElementById("edit-note-message").value = button.dataset.message;
	});
});
