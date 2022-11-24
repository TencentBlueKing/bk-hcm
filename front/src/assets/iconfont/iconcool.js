!(function () {
    var svgCode = '<svg xmlns="http://www.w3.org/2000/svg" data-name="hcm" xmlns:xlink="http://www.w3.org/1999/xlink" style="position:absolute;width:0;height:0;visibility:hidden"><symbol id="bkhcm-icon-automatic-typesetting" viewBox="0 0 64 64"><path d="M57.9 8H6.1c-1.1 0-2 .9-2 2v36c0 1.1.9 2 2 2H30v3.9H12.2v4H52v-4H34V48h23.9c1.1 0 2-.9 2-2V10c0-1.1-.9-2-2-2zm-2.1 36H8.1V12.1h47.7V44z"/><path d="M16.7 39h9.7V17.1H12.2V39h4.5zm5.2-17.4v12.9h-5.2V21.6h5.2zM30.6 34.5H43V39H30.6zM30.6 25.8h20.7v4.5H30.6zM30.6 17.1h20.7v4.5H30.6z"/></g><g fill="#828D97"><path d="M50 46.4H38.4L19.4 30H4v4h13.9l19 16.4H50v7.8l10.1-10.1L50 38zM34.6 20.1H50v8.1l10.1-10.1L50 8v8.1H32.8L22.1 28.2l3 2.6z"/></g></symbol><symbol id="bkhcm-icon-down-shape" viewBox="0 0 24 24"><path d="M12 18a.584.584 0 01-.435-.195l-8.355-9c-.195-.21-.255-.54-.165-.825s.33-.48.6-.48h16.71c.27 0 .495.195.6.48s.03.615-.165.825l-8.355 9A.584.584 0 0112 18z"/></symbol></svg>'
    if (document.body) {
        document.body.insertAdjacentHTML('afterbegin', svgCode)
    } else {
        document.addEventListener('DOMContentLoaded', function() {
            document.body.insertAdjacentHTML('afterbegin', svgCode)
        })
    }
})()