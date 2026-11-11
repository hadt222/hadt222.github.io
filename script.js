const menuToggle = document.getElementById("menu-toggle");
const navList = document.getElementById("nav-list");
const themeToggle = document.getElementById("theme-toggle");
const themeIcon = document.getElementById("theme-icon");
const year = document.getElementById("year");

function updateThemeToggle() {
  if (!themeToggle || !themeIcon) {
    return;
  }

  const isDark = document.documentElement.dataset.theme === "dark";
  themeToggle.setAttribute("aria-label", isDark ? "Switch to light mode" : "Switch to dark mode");
  themeToggle.setAttribute("aria-pressed", String(isDark));
  themeToggle.title = isDark ? "Switch to light mode" : "Switch to dark mode";
  themeIcon.textContent = isDark ? "☀" : "☾";
}

if (themeToggle) {
  updateThemeToggle();

  themeToggle.addEventListener("click", () => {
    const isDark = document.documentElement.dataset.theme === "dark";
    const nextTheme = isDark ? "light" : "dark";

    document.documentElement.dataset.theme = nextTheme;
    localStorage.setItem("theme", nextTheme);
    updateThemeToggle();
  });
}

if (year) {
  year.textContent = new Date().getFullYear();
}

if (menuToggle && navList) {
  menuToggle.addEventListener("click", () => {
    const isOpen = navList.classList.toggle("show");
    menuToggle.setAttribute("aria-expanded", String(isOpen));
    menuToggle.setAttribute("aria-label", isOpen ? "Close menu" : "Open menu");
  });

  navList.querySelectorAll("a").forEach((link) => {
    link.addEventListener("click", () => {
      navList.classList.remove("show");
      menuToggle.setAttribute("aria-expanded", "false");
      menuToggle.setAttribute("aria-label", "Open menu");
    });
  });
}

const observer = new IntersectionObserver(
  (entries) => {
    entries.forEach((entry) => {
      if (entry.isIntersecting) {
        entry.target.classList.add("show");
      }
    });
  },
  { threshold: 0.12 }
);

document.querySelectorAll(".card, .timeline-item, .section-title").forEach((el) => {
  el.classList.add("reveal");
  observer.observe(el);
});