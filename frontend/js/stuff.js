$( document ).ready(function() {
  var wizard = $(".container").steps({
    headerTag: "h3",
    bodyTag: "section",
    stepsOrientation: "vertical",
    enableKeyNavigation: false,
    enableFinishButton: false
  });

  var steps = [
    {"title": "List flavors", "class": "list-flavors"},
    {"title": "List images", "class": "list-images"},
    {"title": "Create servers", "class": "create-server"},
    {"title": "Create LB", "class": "create-lb"}
  ];

  var commandBox = '<div class="command-box"> \
                      <span class="thing">$</span> \
                      <input class="user-input" autocomplete="false" /> \
                      <div class="log-output"></div> \
                    </div>';

  for (var i = 0; i < steps.length; i++) {
    var info = steps[i]
    wizard.steps("add", {
      title: info.title,
      content: '<div id="console" class="console-box step-'+info.class+'">'+commandBox+'</div>'
    });
  }

  focusLastInput();

  $( ".user-input" ).keydown(function(event) {
    if (event.keyCode == 13) {
      path = "flavors"
      $.get("http://localhost:8080/" + path, function (data) {
        // Inject data into current log box
        $("#console .command-box:last .log-output").html("<pre>" + data + "</pre>")

        // Add new prompt
        $("#console").append(commandBox);

        // Focus new input box
        focusLastInput();
      });

    }
    event.stopPropagation();
  });
});

function focusLastInput() {
  $("#console .command-box:last .user-input").focus();
}
