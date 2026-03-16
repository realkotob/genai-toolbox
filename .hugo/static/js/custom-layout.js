document.addEventListener('DOMContentLoaded', function() {

  // Setup CSS for the wrapper and the banner
  var styleTag = document.createElement('style');
  styleTag.innerHTML = `
    .td-navbar .dropdown-menu {
      z-index: 9999 !important;
    }

    .theme-banner-wrapper {
      position: sticky;
      z-index: 20;
      padding-top: 15px;
      padding-bottom: 5px;
      margin-bottom: 2rem;
      background-color: var(--bs-body-bg, #ffffff);
    }

    .theme-migration-banner {
      background-color: #ebf3fc;
      border: 1px solid #80a7e9;
      color: #1c3a6b;
      border-radius: 4px;
      padding: 15px;
      text-align: center;
      width: 100%;
      box-shadow: 0 4px 6px rgba(0,0,0,0.05);
    }

    .theme-migration-banner a {
      color: #4484f4;
      text-decoration: underline;
      font-weight: bold;
    }

    /* DARK MODE STYLING */
    html[data-bs-theme="dark"] .theme-banner-wrapper,
    body.dark .theme-banner-wrapper,
    html.dark-mode .theme-banner-wrapper {
      background-color: var(--bs-body-bg, #20252b);
    }

    html[data-bs-theme="dark"] .theme-migration-banner,
    body.dark .theme-migration-banner,
    html.dark-mode .theme-migration-banner {
      background-color: #1a273b;
      color: #e6efff;
      box-shadow: 0 4px 6px rgba(0,0,0,0.3);
    }

    html[data-bs-theme="dark"] .theme-migration-banner a,
    body.dark .theme-migration-banner a,
    html.dark-mode .theme-migration-banner a {
      color: #80a7e9;
    }

    @media (prefers-color-scheme: dark) {
      html:not([data-bs-theme="light"]):not(.light) .theme-banner-wrapper {
        background-color: var(--bs-body-bg, #20252b);
      }
      html:not([data-bs-theme="light"]):not(.light) .theme-migration-banner {
        background-color: #1a273b;
        color: #e6efff;
        box-shadow: 0 4px 6px rgba(0,0,0,0.3);
      }
      html:not([data-bs-theme="light"]):not(.light) .theme-migration-banner a {
        color: #80a7e9;
      }
    }
  `;
  document.head.appendChild(styleTag);

  // --- DYNAMIC OFFSET CALCULATOR ---
  function updateHeaderOffset() {
    var mainNav = document.querySelector('.td-navbar');
    var secondaryNav = document.getElementById('secondary-nav');
    var migrationWrapper = document.getElementById('migration-banner-wrapper');
    
    var h1 = mainNav ? mainNav.offsetHeight : 0;
    var h2 = secondaryNav ? secondaryNav.offsetHeight : 0;
    var totalHeight = h1 + h2;

    // 1. Update CSS variable for anchor jumps (scroll-margin-top)
    document.documentElement.style.setProperty('--header-offset', totalHeight + 'px');

    // 2. Update the sticky banner position
    if (migrationWrapper) {
      migrationWrapper.style.top = totalHeight + 'px';
    }
  }

  // Create the Wrapper
  var wrapper = document.createElement('div');
  wrapper.id = 'migration-banner-wrapper';
  wrapper.className = 'theme-banner-wrapper';

  // Create the Banner
  var banner = document.createElement('div');
  banner.className = 'theme-migration-banner';
  banner.innerHTML = '⚠️ <strong>Archived Docs:</strong> Visit <a href="https://mcp-toolbox.dev/">mcp-toolbox.dev</a> for the latest version.';
  wrapper.appendChild(banner);

  // Inject the wrapper into the center information column
  var contentArea = document.querySelector('.td-content') || document.querySelector('main');
  if (contentArea) {
    contentArea.prepend(wrapper);
  } else {
    console.warn("Could not find the main content column to inject the banner.");
  }

  // Initialize the dynamic offset
  updateHeaderOffset();

  // Re-calculate on window resize
  window.addEventListener('resize', updateHeaderOffset);

  // Use ResizeObserver to detect header height changes (e.g., mobile menu toggle or wrapping)
  if (window.ResizeObserver) {
    const ro = new ResizeObserver(updateHeaderOffset);
    const navToWatch = document.querySelector('.td-navbar');
    const secNavToWatch = document.getElementById('secondary-nav'); // <-- ADDED: Watches secondary nav for wraps
    if (navToWatch) ro.observe(navToWatch);
    if (secNavToWatch) ro.observe(secNavToWatch); // <-- ADDED: Binds the observer
  }

  // --- BREADCRUMBS REPOSITIONING ---
  var breadcrumbs = document.querySelector('.td-breadcrumbs') || document.querySelector('nav[aria-label="breadcrumb"]');
  var pageTitle = document.querySelector('.td-content h1');

  if (breadcrumbs && pageTitle) {
    pageTitle.parentNode.insertBefore(breadcrumbs, pageTitle);
    breadcrumbs.style.marginTop = "1rem";
    breadcrumbs.style.marginBottom = "2rem";
  }

  // --- DYNAMIC TAB SCROLLING ---
  var tabContainers = document.querySelectorAll('.nav-tabs');

  tabContainers.forEach(function(container) {
    function scrollToCenter(element) {
      if (element) {
        element.scrollIntoView({ behavior: 'smooth', block: 'nearest', inline: 'center' });
      }
    }

    setTimeout(function() {
      var activeTab = container.querySelector('.nav-link.active') || container.querySelector('.active');
      scrollToCenter(activeTab);
    }, 150);

    container.addEventListener('click', function(e) {
      var clickedTab = e.target.closest('a, button, .nav-link');
      if (clickedTab) {
        setTimeout(function() {
          scrollToCenter(clickedTab);
        }, 50);
      }
    });
  });

  // --- HIGHLIGHT ACTIVE VERSION IN DROPDOWN ---
  // <-- ADDED: Finds the active version text and highlights it in the list
  var dropdownToggles = document.querySelectorAll('.td-navbar .dropdown-toggle:not(.td-light-dark-menu)');
  
  dropdownToggles.forEach(function(toggle) {
    var currentText = toggle.textContent.trim();
    var menuItems = toggle.nextElementSibling ? toggle.nextElementSibling.querySelectorAll('.dropdown-item') : [];
    
    menuItems.forEach(function(item) {
      if (item.textContent.trim() === currentText) {
        item.classList.add('active-version');
      }
    });
  });

});