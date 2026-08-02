document.addEventListener("DOMContentLoaded", () => {

        /* Edit */
        const editModal = document.getElementById("editShipperModal");

        editModal.addEventListener("show.bs.modal", function (event) {
                const button = event.relatedTarget;
                document.getElementById("edit-id").value = button.dataset.id;
                document.getElementById("edit-name").value = button.dataset.name;
                document.getElementById("edit-phone").value = button.dataset.phone;
                document.getElementById("edit-email").value = button.dataset.email;
        });

        /* Delete */
        const deleteModal = document.getElementById("deleteShipperModal");

        deleteModal.addEventListener("show.bs.modal", function (event) {
                const button = event.relatedTarget;
                document.getElementById("delete-id").value = button.dataset.id;
                document.getElementById("delete-name").textContent = button.dataset.name;
        });
});
