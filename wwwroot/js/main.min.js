(() => {
  'use strict';

  const navbar = document.querySelector('.navbar');
  const emailInput = document.getElementById('email-input');
  const notifyBtn = document.getElementById('notify-button');
  const emailRE = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

  // Smooth scroll for anchor links
  document.addEventListener(
    'click',
    (e) => {
      const a = e.target.closest('a[href^="#"]');
      if (!a) return;

      const href = a.getAttribute('href');
      if (!href || href === '#') return;

      const target = document.querySelector(href);
      if (!target) return;

      e.preventDefault();
      target.scrollIntoView({ behavior: 'smooth', block: 'start' });
    },
    { passive: false }
  );

  // Intersection Observer for fade-in animations
  if ('IntersectionObserver' in window) {
    const observer = new IntersectionObserver(
      (entries, obs) => {
        for (const entry of entries) {
          if (entry.isIntersecting) {
            entry.target.classList.add('is-visible');
            obs.unobserve(entry.target);
          }
        }
      },
      { threshold: 0.1, rootMargin: '0px 0px -50px 0px' }
    );

    document.querySelectorAll('.fade-in').forEach((el) => {
      observer.observe(el);
    });
  } else {
    document.querySelectorAll('.fade-in').forEach((el) => el.classList.add('is-visible'));
  }

  // Newsletter card animations on load
  window.addEventListener('load', () => {
    document
      .querySelectorAll('.newsletter-header, .newsletter-content')
      .forEach((el, i) => {
        el.style.transitionDelay = `${i * 0.2}s`;
        el.classList.add('is-visible');
      });

    document.querySelectorAll('.newsletter-card').forEach((card, i) => {
      card.style.transitionDelay = `${i * 0.1}s`;
      card.classList.add('is-visible');
    });
  });

  // Navbar scroll effect
  let lastY = 0;
  let ticking = false;

  function onScroll() {
    lastY = window.scrollY || window.pageYOffset;
    if (!ticking) {
      requestAnimationFrame(updateOnScroll);
      ticking = true;
    }
  }

  function updateOnScroll() {
    if (navbar) {
      navbar.classList.toggle('scrolled', lastY > 50);
    }

    const progressBar = document.querySelector('.reading-progress');
    if (progressBar) {
      const max = document.body.scrollHeight - window.innerHeight;
      const progress = max > 0 ? Math.min(Math.max(lastY / max, 0), 1) : 0;
      progressBar.style.width = `${progress * 100}%`;
    }

    ticking = false;
  }

  window.addEventListener('scroll', onScroll, { passive: true });
  updateOnScroll();

  // Subscribe handler (hero + CTA)
  function setupSubscribe(inputEl, btnEl) {
    if (!btnEl || !inputEl) return;

    let inFlight = false;
    const controller = new AbortController();

    btnEl.addEventListener('click', async () => {
      const email = inputEl.value.trim();
      if (!emailRE.test(email)) {
        alert('Please enter a valid email address.');
        inputEl.focus();
        return;
      }
      if (inFlight) return;

      inFlight = true;
      btnEl.disabled = true;

      try {
        const res = await fetch('/subscribe', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ email }),
          signal: controller.signal,
        });

        let message = 'Thank you for subscribing!';
        if (res.ok) {
          const data = await res.json().catch(() => ({}));
          message = data.message || message;
        } else {
          message = "Thanks! We'll notify you when we launch.";
        }

        alert(message);
        inputEl.value = '';
      } catch (err) {
        console.error('Subscribe error:', err);
        alert("Thanks! We'll notify you when we launch.");
        inputEl.value = '';
      } finally {
        btnEl.disabled = false;
        inFlight = false;
      }
    });

    window.addEventListener('beforeunload', () => controller.abort(), {
      passive: true,
    });
  }

  // Setup both hero and CTA subscribe forms
  setupSubscribe(emailInput, notifyBtn);
  setupSubscribe(
    document.getElementById('cta-email-input'),
    document.getElementById('cta-notify-button')
  );

  // Share newsletter utility
  window.shareNewsletter = async function shareNewsletter() {
    try {
      if (navigator.share) {
        await navigator.share({
          title: document.title,
          text: 'Check out this newsletter from RideAware',
          url: location.href,
        });
        return;
      }
    } catch (err) {
      console.warn('navigator.share error/cancel:', err);
    }

    if (navigator.clipboard && window.isSecureContext) {
      try {
        await navigator.clipboard.writeText(location.href);
        alert('Newsletter URL copied to clipboard!');
        return;
      } catch {
        /* fall through */
      }
    }

    const tmp = document.createElement('input');
    tmp.value = location.href;
    document.body.appendChild(tmp);
    tmp.select();
    document.execCommand('copy');
    document.body.removeChild(tmp);
    alert('Newsletter URL copied to clipboard!');
  };
})();
