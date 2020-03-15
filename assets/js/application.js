require("expose-loader?$!expose-loader?jQuery!jquery");
import "bootstrap/dist/js/bootstrap.bundle"


marked.setOptions({
  renderer: new marked.Renderer(),
  gfm: true,
  tables: true,
  breaks: false,
  pedantic: false,
  sanitize: false,
  smartLists: true,
  smartypants: false
});

$(function () {
  $("#post-Content").keyup(function (e) {
    var text = $(e.target).val();
    $("#content-preview").html(marked(text));
  });
});
