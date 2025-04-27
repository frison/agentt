(() => {
  const themeToggle = document.getElementById('theme-toggle');
  const systemPrefersDark = window.matchMedia('(prefers-color-scheme: dark)');
  const htmlElement = document.documentElement; // Target the <html> element

  // Function to set the theme
  const setTheme = (theme) => {
    htmlElement.setAttribute('data-theme', theme); // Set on <html>
    localStorage.setItem('theme', theme);
    // Update icon visibility is handled by CSS based on data-theme
  };

  // Function to toggle the theme
  const toggleTheme = () => {
    const currentTheme = htmlElement.getAttribute('data-theme'); // Read from <html>
    const newTheme = currentTheme === 'dark' ? 'light' : 'dark';
    setTheme(newTheme);
  };

  // Initial theme is already set by the inline script in <head>
  // We just need to ensure the button listener is added.

  // Add event listener for the toggle button
  if (themeToggle) {
    themeToggle.addEventListener('click', toggleTheme);
  }

  // Optional: Listen for changes in system preference while the page is open
  // Update this listener to also only apply if localStorage is not set,
  // matching the behavior of the inline script.
  systemPrefersDark.addEventListener('change', (e) => {
    if (!localStorage.getItem('theme')) {
       setTheme(e.matches ? 'dark' : 'light');
    }
  });

})();