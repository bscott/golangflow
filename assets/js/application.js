require('expose-loader?$!expose-loader?jQuery!jquery');
require("bootstrap/dist/js/bootstrap.js");
let marked = require("marked/lib/marked.js");

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

$(() => {

  $("#post-Content").keyup((e)=>{
    let text = $(e.target).val();
    $("#content-preview").html(marked(text));
  })

  $("#show-preview").keyup((e)=>{
    let text = $(e.target).val();
    $("#post-preview").html(marked(text));
  })

});



