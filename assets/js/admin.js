document.addEventListener('DOMContentLoaded', function () {
  document.addEventListener('submit', function (e) {
    var msg = e.target.dataset.confirm;
    if (msg && !window.confirm(msg)) {
      e.preventDefault();
    }
  });
});
