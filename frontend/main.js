var provUrl = location.protocol + "//" + location.hostname + ':8080';

var commandBox = '<div class="command-box"> \
                    <span class="thing">$</span> \
                    <input class="user-input" autocomplete="false" /> \
                    <div class="log-output"></div> \
                  </div>';

$( document ).ready(function() {

  hideWarning();
  focusLastInput();
  progressStep(1);

  var serverIPs = [];

  $( document ).on("keydown", ".user-input", function(event) {
    if (event.keyCode == 13) {

      var cmd = $(this).val();
      var serverCmdPatt = /^server create (.+)$/;
      var lbCreatePatt = /^lb create (.+)$/;

      // Execute appropriate command
      if (cmd == "help") {
        populateOutput(helpText);
      } else if (cmd == "clear") {
        appendNewCommandBox();
        $("#console .command-box:not(:last)").hide();
      } else if (cmd == "flavor list") {
          hideWarning();
          $.get(provUrl+"/flavors", function (data) {
            populateOutput(data);
            appendNewCommandBox();
            progressStep(2);
          });
      } else if (cmd == "image list") {
          hideWarning();
          $.get(provUrl+"/images", function (data) {
            populateOutput(data);
            appendNewCommandBox();
            progressStep(3);
          });
      } else if (serverCmdPatt.test(cmd)) {
        if (serverIPs.length >= 2) {
          triggerWarning("Whoa there, Nelly! You're only allowed to provision 2 servers.");
          return false;
        }
        var match = serverCmdPatt.exec(cmd)
        if (match == null || match[1] == undefined) {
          triggerWarning("The name for your server was not understood. Please provide another name.");
          return false;
        }
        var json = JSON.stringify({name: match[1]});
        hideWarning();
        $.post(provUrl+"/create_server", json, function (data, txt, xhr) {
            populateOutput(data);
            appendNewCommandBox();
            serverIPs.push(xhr.getResponseHeader('Server-Ip'));
            progressStep(4);
        });
      } else if (lbCreatePatt.test(cmd)) {
        var match = lbCreatePatt.exec(cmd);
        if (match == null || match[1] == undefined) {
          triggerWarning("The name for your LB was not understood. Please provide another name.");
          return false;
        }
        var json = JSON.stringify({name: match[1], server_ips: serverIPs})
        hideWarning();
        $.post(provUrl+"/create_lb", json, function (data, txt, xhr) {
            populateOutput(data);
            appendNewCommandBox();
            progressStep(5);
        });
      } else {
          triggerWarning('"'+cmd+'" is not a supported command. Try again ;)');
          appendNewCommandBox();
      }
    }

    event.stopPropagation();
  });

  $(document).on("click", ".console-box", function() {
    // If the user is selecting text, don't fire focus
    if (!getSelection().toString()) {
      focusLastInput();
    }
  })
});

function progressStep(step) {
  $(".steps li.current").removeClass("current").addClass("done");

  var messages = {
      1: "Welcome to the Cloud Console tour! The aim of this site is to provide \
          a quick introduction to the services that Rackspace offers, for free! \
          This tour is divided into four steps: first we will select a hardware \
          flavor, then select an operating system, then provision a few servers \
          and then spin up a load balancer.<br><br>So let's begin. To browse all \
          available flavors, run this command: <strong>flavor list</strong>",
      2: "Success! Listed below are all the available hardware configurations, \
          or \"flavors\", that servers can be based on. Each flavor has a \
          particular variety of RAM and disk capacity - allowing you to pick the \
          configuration most suited to your workflow.<br><br>Next we'll look at \
          all the available operating systems you can use. To list them, just run \
          this command: <strong>image list</strong>",
      3: "Cool - listed below are all the available operating systems our server \
          can use. Now that we know about flavors and images, we're ready to \
          provision a VM! For the sake of this tour, we will select \
          a particular flavor and image for you. Your server will have 1GB of RAM \
          and use the Ubuntu 12.04 operating system - all you need to do is pick a \
          name.<br><br>To do this, run this command: <strong>server create {name}</strong> \
          , where {name} is the name you want to use.",
      4: "Now that we have our first server, let's create another one. Do the same \
          as before, but choose a different name: <strong>server create {name} \
          </strong>. When you're done, feel free to SSH into each VM to check \
          they're there. Run this command from your own shell: <strong>ssh root@\
          {ip_address}</strong> using the passwords provided.",
      5: "So we have two servers, great. Now we want to provision a Load Balancer \
          to evenly distribute incoming traffic between them. <br><br>To do this, all we \
          need to do is select a name and run: <strong>lb create {name}</strong>.",
      6: "Awesome! You've finished. Visit the link returned below and you'll see \
          the front-end"
  };

  if (step == 6) {
    $("#info").hide();
    $("#success").html(messages[6]).show();
  } else {
    $(".steps .step" + step).removeClass("disabled done").addClass("current");
    if (messages[step] != undefined) {
      $("#info").html(messages[step]).show();
    }
  }
}

function appendNewCommandBox() {
  // Disable this prompt and unfocus
  $(this).attr("disabled", "true").blur();

  // Add new prompt
  $("#console").append(commandBox);

  // Focus new input box
  focusLastInput();
}

function populateOutput(data) {
  // Inject data into current log box
  $("#console .command-box:last .log-output").html("<pre>" + data + "</pre>")
}

function triggerWarning(msg) {
  $("#warning").text(msg).show();
}

function hideWarning() {
  $("#warning").hide();
}

function focusLastInput() {
  $("#console .command-box:last .user-input").focus();
}
