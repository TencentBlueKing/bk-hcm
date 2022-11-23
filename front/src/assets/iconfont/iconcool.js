!(function () {
    var svgCode = '<svg xmlns="http://www.w3.org/2000/svg" data-name="hcm" xmlns:xlink="http://www.w3.org/1999/xlink" style="position:absolute;width:0;height:0;visibility:hidden"><symbol id="bkhcm-icon-down-shape" viewBox="0 0 24 24"><path d="M12 18a.584.584 0 01-.435-.195l-8.355-9c-.195-.21-.255-.54-.165-.825s.33-.48.6-.48h16.71c.27 0 .495.195.6.48s.03.615-.165.825l-8.355 9A.584.584 0 0112 18z"/></symbol></svg>'
    if (document.body) {
        document.body.insertAdjacentHTML('afterbegin', svgCode)
    } else {
        document.addEventListener('DOMContentLoaded', function() {
            document.body.insertAdjacentHTML('afterbegin', svgCode)
        })
    }
})()