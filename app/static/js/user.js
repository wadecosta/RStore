document.addEventListener("DOMContentLoaded", () => {
	const editModal = document.getElementById("editAccountModal");

	editModal.addEventListener("show.bs.modal", function (event) {
		const button = event.relatedTarget;
		document.getElementById("user-edit-username").value = button.dataset.username;
		document.getElementById("user-edit-email").value = button.dataset.email;
	});
});
