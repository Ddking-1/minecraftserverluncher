(function() {
    const themeToggle = document.getElementById('themeToggle');
    const themeIcon = document.getElementById('themeIcon');

    function getThemeIconPath(theme) {
        if (theme === 'dark') {
            return 'M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z';
        } else {
            return 'M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z';
        }
    }

    function updateThemeIcon() {
        const theme = document.documentElement.getAttribute('data-theme');
        themeIcon.setAttribute('d', getThemeIconPath(theme));
    }

    function setTheme(theme) {
        document.documentElement.setAttribute('data-theme', theme);
        localStorage.setItem('theme', theme);
        updateThemeIcon();
    }

    function toggleTheme() {
        const current = document.documentElement.getAttribute('data-theme');
        setTheme(current === 'dark' ? 'light' : 'dark');
    }

    themeToggle.addEventListener('click', toggleTheme);
    updateThemeIcon();

    window.setTheme = setTheme;
    window.toggleTheme = toggleTheme;
})();
