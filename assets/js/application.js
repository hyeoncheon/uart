require("expose-loader?$!expose-loader?jQuery!jquery");
require("bootstrap/dist/js/bootstrap.bundle.js");
require("@fortawesome/fontawesome-free/js/all.js");
require("jquery-ujs/src/rails.js");

$(() => {
	// enabling bootstrap widgets
	$('[data-toggle="popover"]').popover();
	$('[data-toggle="tooltip"]').tooltip();

	// auto-close alerts
	window.setTimeout(function() {
		$(".alert:not('.alert-danger')").alert('close');
	}, 10000);

	// navigation position highlighter
	var current_path = document.location.pathname;
	$(".nav-item").removeClass("active");
	$(".nav-item").each(function(index) {
		if ($(this).attr('href') == current_path) {
			$(this).addClass("active");
			return false; // exit the loop
		}
	});
	$(".dropdown-item").removeClass("active");
	$(".dropdown-item").each(function(index) {
		if ($(this).attr('href') == current_path) {
			$(this).addClass("active");
			$(this).parent().parent().addClass("active");
			return false; // exit the loop
		}
	});

	// table row with link
	$('tr.linked > td[class!="nolink"]').click(function() {
		window.location = $(this).parent().attr('target');
	});

	// use moment for time fields
	moment.locale(navigator.language);
	$('.moment').each(function(i, e) {
		var format = $(e).attr('form');
		if (format == undefined) {
			format = "YYYY-MM-DD hh:mm";
		}
		var time = moment($(e).text());
		var disp = time.format(format);
		if (moment().diff(time, 'months') < 1 && !$(e).hasClass("norel")) {
			disp = time.fromNow();
		}
		$(e).html('<span title="' + time.format() + '">' + disp + '</span>');
	});

	// EasyMDE: https://github.com/Ionaru/easy-markdown-editor
	var easyMDE = new EasyMDE({
		element: document.getElementById("doc-Content"),
		autoDownloadFontAwesome: false,
		autosave: {
			enabled: false,
			uniqueId: "doc-content",
		},
		lineWrapping: true,
		renderingConfig: {
			singleLineBreaks: false,
			codeSyntaxHighlighting: true,
		},
		spellChecker: false,
	});

	// highlight.js, see https://highlightjs.org/usage/
	$('.highlight pre').each(function(i, block) {
		hljs.highlightBlock(block);
	});

	// goback
	$('a.goback').on('click', function(e){
		e.preventDefault();
		window.history.back();
	});

});
