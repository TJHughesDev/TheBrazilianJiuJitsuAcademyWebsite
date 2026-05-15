(function () {
  "use strict";

  /* ─── Utility ───────────────────────────────────────────────── */
  function escHtml(str) {
    return String(str)
      .replace(/&/g, "&amp;")
      .replace(/</g, "&lt;")
      .replace(/>/g, "&gt;")
      .replace(/"/g, "&quot;")
      .replace(/'/g, "&#39;");
  }

  /* ─── Nav scroll ────────────────────────────────────────────── */
  var nav = document.getElementById("nav");
  function onScroll() {
    nav.classList.toggle("scrolled", window.scrollY > 40);
  }
  window.addEventListener("scroll", onScroll, { passive: true });
  onScroll();

  /* ─── Mobile menu ───────────────────────────────────────────── */
  var hamburger     = document.getElementById("hamburger");
  var mobileMenu    = document.getElementById("mobileMenu");
  var mobileOverlay = document.getElementById("mobileOverlay");
  var menuClose     = document.getElementById("menuClose");
  var mobileLinks   = document.querySelectorAll(".mobile-link");

  function openMenu() {
    mobileMenu.classList.add("open");
    mobileOverlay.classList.add("open");
    mobileMenu.setAttribute("aria-hidden", "false");
    hamburger.setAttribute("aria-expanded", "true");
    document.body.style.overflow = "hidden";
  }

  function closeMenu() {
    mobileMenu.classList.remove("open");
    mobileOverlay.classList.remove("open");
    mobileMenu.setAttribute("aria-hidden", "true");
    hamburger.setAttribute("aria-expanded", "false");
    document.body.style.overflow = "";
  }

  hamburger.addEventListener("click", openMenu);
  menuClose.addEventListener("click", closeMenu);
  mobileOverlay.addEventListener("click", closeMenu);
  mobileLinks.forEach(function (el) { el.addEventListener("click", closeMenu); });
  document.addEventListener("keydown", function (e) { if (e.key === "Escape") closeMenu(); });

  /* ─── Scroll reveal ─────────────────────────────────────────── */
  var revealEls = document.querySelectorAll(".reveal");
  if ("IntersectionObserver" in window) {
    var revealObserver = new IntersectionObserver(function (entries) {
      entries.forEach(function (entry) {
        if (entry.isIntersecting) {
          entry.target.classList.add("in-view");
          revealObserver.unobserve(entry.target);
        }
      });
    }, { threshold: 0.12, rootMargin: "0px 0px -40px 0px" });
    revealEls.forEach(function (el) { revealObserver.observe(el); });
  } else {
    revealEls.forEach(function (el) { el.classList.add("in-view"); });
  }

  /* ─── FAQ accordion ─────────────────────────────────────────── */
  var faqItems = document.querySelectorAll(".faq__item");
  faqItems.forEach(function (item) {
    var btn = item.querySelector(".faq__q");
    var ans = item.querySelector(".faq__a");
    btn.addEventListener("click", function () {
      var isOpen = btn.getAttribute("aria-expanded") === "true";
      faqItems.forEach(function (other) {
        other.querySelector(".faq__q").setAttribute("aria-expanded", "false");
        other.querySelector(".faq__a").classList.remove("open");
      });
      if (!isOpen) {
        btn.setAttribute("aria-expanded", "true");
        ans.classList.add("open");
      }
    });
  });

  /* ─── Testimonials ──────────────────────────────────────────── */
  var carouselTrack = document.getElementById("carouselTrack");

  if (carouselTrack) {
    carouselTrack.addEventListener("touchstart", function () {
      carouselTrack.style.animationPlayState = "paused";
    }, { passive: true });
    carouselTrack.addEventListener("touchend", function () {
      carouselTrack.style.animationPlayState = "";
    }, { passive: true });

    fetch("/data/testimonials.json")
      .then(function (r) { return r.json(); })
      .then(function (testimonials) {
        if (!testimonials || !testimonials.length) return;

        var html = testimonials.map(function (t) {
          var stars  = "★".repeat(t.rating);
          var avatar = t.image
            ? '<img src="' + escHtml(t.image) + '" alt="' + escHtml(t.name) + '" loading="lazy">'
            : '<svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true"><path d="M12 12c2.761 0 5-2.239 5-5s-2.239-5-5-5-5 2.239-5 5 2.239 5 5 5zm0 2c-3.337 0-10 1.676-10 5v2h20v-2c0-3.324-6.663-5-10-5z"/></svg>';
          return (
            '<div class="testimonial">' +
              '<div class="testimonial__stars">' + stars + "</div>" +
              '<p class="testimonial__quote">"' + escHtml(t.quote) + '"</p>' +
              '<div class="testimonial__author">' +
                '<div class="testimonial__avatar">' + avatar + "</div>" +
                "<span>" + escHtml(t.name) + "</span>" +
              "</div>" +
            "</div>"
          );
        }).join("");

        // Duplicate for seamless infinite marquee
        carouselTrack.innerHTML = html + html;
      });
  }

  /* ─── Schedule (home page) ──────────────────────────────────── */
  var scheduleGrid = document.getElementById("scheduleGrid");

  if (scheduleGrid) {
    var DAY_ABBR = {
      Monday: "MON", Tuesday: "TUE", Wednesday: "WED",
      Thursday: "THU", Friday: "FRI", Saturday: "SAT", Sunday: "SUN",
    };

    function typeBadge(type) {
      if (type === "Kids")   return '<span class="schedule-card__badge badge--kids">Kids</span>';
      if (type === "Adults") return '<span class="schedule-card__badge badge--adults">Adults</span>';
      return '<span class="schedule-card__badge badge--all">All Levels</span>';
    }

    fetch("/data/schedule.json")
      .then(function (r) { return r.json(); })
      .then(function (sessions) {
        if (!sessions || !sessions.length) {
          scheduleGrid.innerHTML = '<p style="color:var(--grey-3);padding:32px 0;">No sessions listed yet.</p>';
          return;
        }
        scheduleGrid.innerHTML = sessions.map(function (s) {
          return (
            '<div class="schedule-card">' +
              '<div class="schedule-card__day">' +
                '<span class="schedule-card__day-abbr">' + (DAY_ABBR[s.day] || s.day.slice(0, 3).toUpperCase()) + "</span>" +
              "</div>" +
              '<div class="schedule-card__info">' +
                '<div class="schedule-card__name">' + escHtml(s.class) + "</div>" +
                '<div class="schedule-card__time">' + escHtml(s.time) + "</div>" +
                typeBadge(s.type) +
              "</div>" +
            "</div>"
          );
        }).join("");
      })
      .catch(function () {
        scheduleGrid.innerHTML = '<p style="color:var(--grey-3);padding:32px 0;">Could not load schedule. Please contact us directly.</p>';
      });
  }

  /* ─── Instructors (home page) ───────────────────────────────── */
  var instructorsGrid = document.getElementById("instructorsGrid");

  if (instructorsGrid) {
    var BELT_CLASS = {
      "Black Belt": "belt--black", "Brown Belt": "belt--brown",
      "Purple Belt": "belt--purple", "Blue Belt": "belt--blue", "White Belt": "belt--white",
    };

    function initials(name) {
      return name.split(" ").map(function (w) { return w[0]; }).join("").slice(0, 2).toUpperCase();
    }

    fetch("/data/instructors.json")
      .then(function (r) { return r.json(); })
      .then(function (instructors) {
        if (!instructors || !instructors.length) {
          instructorsGrid.innerHTML = '<p style="color:var(--grey-3);padding:32px 0;">No instructors listed yet.</p>';
          return;
        }
        instructorsGrid.innerHTML = instructors.map(function (i) {
          var beltClass  = BELT_CLASS[i.belt] || "belt--black";
          var lineageHtml = i.lineage
            ? '<div class="instructor-card__lineage"><strong>Lineage</strong>' + escHtml(i.lineage) + "</div>"
            : "";
          return (
            '<div class="instructor-card">' +
              '<div class="instructor-card__avatar">' +
                (i.image ? '<img src="' + escHtml(i.image) + '" alt="' + escHtml(i.name) + '" loading="lazy">' : initials(i.name)) +
              "</div>" +
              '<span class="instructor-card__belt ' + beltClass + '">' + escHtml(i.belt) + "</span>" +
              '<div class="instructor-card__name">' + escHtml(i.name) + "</div>" +
              '<div class="instructor-card__title">' + escHtml(i.title) + "</div>" +
              '<p class="instructor-card__bio">' + escHtml(i.bio) + "</p>" +
              lineageHtml +
            "</div>"
          );
        }).join("");
      })
      .catch(function () {
        instructorsGrid.innerHTML = '<p style="color:var(--grey-3);padding:32px 0;">Could not load instructor information.</p>';
      });
  }

  /* ─── Timetable (timetable page) ────────────────────────────── */
  var ttGrid = document.getElementById("ttGrid");

  if (ttGrid) {
    fetch("/data/timetable.json")
      .then(function (r) { return r.json(); })
      .then(function (days) {
        ttGrid.innerHTML = days.map(function (day) {
          var hasSessions = day.sessions && day.sessions.length;
          var sessionsHtml = hasSessions
            ? '<div class="tt-day__sessions">' +
                day.sessions.map(function (s) {
                  return (
                    '<div class="tt-session tt-session--' + escHtml(s.type) + '">' +
                      '<div class="tt-session__time">' + escHtml(s.time) + "</div>" +
                      '<div class="tt-session__name">' + escHtml(s.name) + "</div>" +
                    "</div>"
                  );
                }).join("") +
              "</div>"
            : "";
          return (
            '<div class="tt-day' + (hasSessions ? "" : " tt-day--rest") + '">' +
              '<div class="tt-day__header">' +
                '<span class="tt-day__name">' + escHtml(day.day) + "</span>" +
                (!hasSessions ? '<span class="tt-day__rest-label">No scheduled classes</span>' : "") +
              "</div>" +
              sessionsHtml +
            "</div>"
          );
        }).join("");
      })
      .catch(function () {
        ttGrid.innerHTML = '<p style="color:var(--grey-3);padding:32px 0;">Could not load timetable. Please contact us directly.</p>';
      });
  }

  /* ─── Taster form (home page) ───────────────────────────────── */
  var form = document.getElementById("tasterForm");

  if (form) {
    var submitBtn   = document.getElementById("submitBtn");
    var formSuccess = document.getElementById("formSuccess");
    var formError   = document.getElementById("formError");

    function setLoading(yes) {
      submitBtn.disabled = yes;
      submitBtn.classList.toggle("loading", yes);
    }

    function showSuccess(msg) {
      var p = formSuccess.querySelector("p");
      if (p && msg) p.textContent = msg;
      formSuccess.classList.remove("hidden");
      formError.classList.add("hidden");
    }

    function showError() {
      formError.classList.remove("hidden");
      formSuccess.classList.add("hidden");
    }

    function validateForm() {
      var ok = true;
      ["name", "email"].forEach(function (field) {
        var el = form.elements[field];
        if (!el) return;
        if (!el.value.trim()) { el.classList.add("error"); ok = false; }
        else { el.classList.remove("error"); }
      });
      var emailEl = form.elements["email"];
      if (emailEl && emailEl.value && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(emailEl.value)) {
        emailEl.classList.add("error");
        ok = false;
      }
      return ok;
    }

    form.addEventListener("submit", function (e) {
      e.preventDefault();
      if (!validateForm()) return;

      var data = {
        name:     form.elements["name"]     ? form.elements["name"].value.trim()    : "",
        email:    form.elements["email"]    ? form.elements["email"].value.trim()   : "",
        phone:    form.elements["phone"]    ? form.elements["phone"].value.trim()   : "",
        interest: form.elements["interest"] ? form.elements["interest"].value       : "",
        message:  form.elements["message"]  ? form.elements["message"].value.trim() : "",
      };

      setLoading(true);
      formSuccess.classList.add("hidden");
      formError.classList.add("hidden");

      fetch("https://formspree.io/f/mdabrrwn", {
        method: "POST",
        headers: { "Content-Type": "application/json", Accept: "application/json" },
        body: JSON.stringify(data),
      })
        .then(function (r) { return r.json(); })
        .then(function (res) {
          setLoading(false);
          if (res.ok) { showSuccess("Thanks! We'll be in touch shortly."); form.reset(); }
          else { showError(); }
        })
        .catch(function () { setLoading(false); showError(); });
    });

    form.addEventListener("input", function (e) {
      if (e.target) e.target.classList.remove("error");
    });
  }

})();
