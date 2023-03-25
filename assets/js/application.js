require("expose-loader?exposes=$,jQuery!jquery");
require("bootstrap/dist/js/bootstrap.bundle.js");
require("@fortawesome/fontawesome-free/js/all.js");
require("jquery-ujs/src/rails.js");

import { marked } from 'marked';



marked.setOptions({
  renderer: new marked.Renderer(),
  gfm: true,
  tables: true,
  breaks: false,
  pedantic: false,
  sanitize: true,
  smartLists: true,
  smartypants: false
});

$(function () {
  $("#post-Content").keyup(function (e) {
    var text = $(e.target).val();
    $("#content-preview").html(marked(text));
  });
});
