require('expose-loader?$!expose-loader?jQuery!jquery');
require("bootstrap/dist/js/bootstrap.js");

$(document).ready(function(){
	// enabling bootstrap widgets
	$('[data-toggle="popover"]').popover();
	$('[data-toggle="tooltip"]').tooltip();

	// navigation position highlighter
	$(".navbar-nav a:not('.selector')").parent().removeClass("active");
	$(".navbar-nav a:not('.selector')").each(function(index) {
		if ($(this).attr('href') == document.location.pathname) {
			$(this).parent().addClass("active");
		}
	});

	// auto-close alerts
	window.setTimeout(function() {
		$(".alert:not('.alert-danger')").alert('close');
	}, 10000);

	$('tr.clickable > td[class!="unclickable"]').click(function() {
		window.location = $(this).parent().find('a#link').attr('href');
	});
});

$(document).ready(function(){
	$('a.goback').on('click', function(e){
		e.preventDefault();
		window.history.back();
	});
});

$(() => {

});
