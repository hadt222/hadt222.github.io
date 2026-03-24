const menuToggle = document.getElementById("menu-toggle");
const navList = document.getElementById("nav-list");
const year = document.getElementById("year");

if (year) {
  year.textContent = new Date().getFullYear();
}

if (menuToggle && navList) {
  menuToggle.addEventListener("click", () => {
    navList.classList.toggle("show");
  });

  navList.querySelectorAll("a").forEach((link) => {
    link.addEventListener("click", () => {
      navList.classList.remove("show");
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