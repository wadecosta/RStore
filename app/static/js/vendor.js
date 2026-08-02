document.addEventListener("DOMContentLoaded", () => {

	/* Edit */
	const editModal = document.getElementById("editVendorModal");
	
	editModal.addEventListener("show.bs.modal", function (event) {
		const button = event.relatedTarget;
		document.getElementById("edit-vendor-id").value = button.dataset.id;
		document.getElementById("edit-vendor-name").value = button.dataset.name;
		document.getElementById("edit-vendor-phone").value = button.dataset.phone;
		document.getElementById("edit-vendor-email").value = button.dataset.email;
		document.getElementById("edit-vendor-auto").checked = button.dataset.auto === "true";
	});

	/* Delete */
	const deleteModal = document.getElementById("deleteVendorModal");

	deleteModal.addEventListener("show.bs.modal", function (event) {
		const button = event.relatedTarget;
		document.getElementById("delete-vendor-id").value = button.dataset.id;
		document.getElementById("delete-vendor-name").textContent = button.dataset.name;
	});
});
