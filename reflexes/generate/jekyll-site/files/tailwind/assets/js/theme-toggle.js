(() => {
  const themeToggle = document.getElementById('theme-toggle');
  const systemPrefersDark = window.matchMedia('(prefers-color-scheme: dark)');
  const htmlElement = document.documentElement; // Target the <html> element

  // Function to set the theme class on <html>
  const setTheme = (theme) => {
    if (theme === 'dark') {
      htmlElement.classList.add('dark');
    } else {
      htmlElement.classList.remove('dark');
    }
    localStorage.setItem('theme', theme);
    // Icon visibility update is needed if not handled by CSS
    updateIconVisibility(theme);
  };

  // Function to toggle the theme
  const toggleTheme = () => {
    // Check if the 'dark' class is currently present
    const isDarkMode = htmlElement.classList.contains('dark');
    const newTheme = isDarkMode ? 'light' : 'dark';
    setTheme(newTheme);
  };

  // Function to update icon visibility (can be done via CSS too)
  const updateIconVisibility = (theme) => {
    const moonIcon = themeToggle?.querySelector('.icon-moon');
    const sunIcon = themeToggle?.querySelector('.icon-sun');
    if (moonIcon && sunIcon) {
      if (theme === 'dark') {
        moonIcon.style.display = 'none';
        sunIcon.style.display = 'inline';
      } else {
        moonIcon.style.display = 'inline';
        sunIcon.style.display = 'none';
      }
    }
  };

  // Initial theme is already set by the inline script in <head>
  // We just need to ensure the button listener is added and initial icon state is correct.

  // Add event listener for the toggle button
  if (themeToggle) {
    themeToggle.addEventListener('click', toggleTheme);
    // Set initial icon visibility based on the theme set by the inline script
    const initialTheme = htmlElement.classList.contains('dark') ? 'dark' : 'light';
    updateIconVisibility(initialTheme);
  }

  // Optional: Listen for changes in system preference
  systemPrefersDark.addEventListener('change', (e) => {
    if (!localStorage.getItem('theme')) { // Only change if no user preference is set
      setTheme(e.matches ? 'dark' : 'light');
    }
  });

})();