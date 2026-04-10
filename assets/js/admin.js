document.addEventListener('DOMContentLoaded', function () {
  // Form confirmation dialogs
  document.addEventListener('submit', function (e) {
    var msg = e.target.dataset.confirm;
    if (msg && !window.confirm(msg)) {
      e.preventDefault();
    }
  });

  // Theme switcher
  var saved = localStorage.getItem('goadmin-theme');
  if (saved) {
    setTheme(saved);
  }
});

function setTheme(theme) {
  document.documentElement.setAttribute('data-theme', theme);
  localStorage.setItem('goadmin-theme', theme);
  var labels = {light: 0, medium: 1, dark: 2};
  var idx = labels[theme];
  document.querySelectorAll('.footer-theme-selector').forEach(function (sel) {
    for (var i = 0; i < sel.children.length; i++) {
      sel.children[i].classList.toggle('active-ft', i === idx);
    }
  });
}
