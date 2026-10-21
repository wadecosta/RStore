document.addEventListener("DOMContentLoaded", () => {

	/* Edit */
	const editModal = document.getElementById("editRecommendationModal");
	
	editModal.addEventListener("show.bs.modal", function (event) {
		const button = event.relatedTarget;
		document.getElementById("recommendation-edit-id").value = button.dataset.id;
		document.getElementById("recommendation-edit-name").value = button.dataset.name;
		document.getElementById("recommendation-edit-link").value = button.dataset.link;
	});
});
