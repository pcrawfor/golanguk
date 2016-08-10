window.addEventListener("load", function(evt) {
    var output = document.getElementById("output");
    var input = document.getElementById("input");
    var ws;
    var print = function(message) {
        var d = document.createElement("div");
        d.innerHTML = message;
        if(output.childNodes.length > 0) {
          output.insertBefore(d, output.childNodes[0])
        } else {
          output.appendChild(d);
        }
    };

    var showGif = function(url) {
        console.log("link");
        var d = document.createElement("div");
        var img = document.createElement("img");
        img.setAttribute("src", url);
        d.appendChild(img);
        if(output.childNodes.length > 0) {
          output.insertBefore(d, output.childNodes[0])
        } else {
          output.appendChild(d);
        }
    };

    document.getElementById("clear").onclick = function(evt) {
        var myNode = document.getElementById("output");
        while (myNode.firstChild) {
          myNode.removeChild(myNode.firstChild);
        }
    };

    document.getElementById("open").onclick = function(evt) {
        if (ws) {
            return false;
        }
        ws = new WebSocket("{{.}}");
        ws.onopen = function(evt) {
            print("OPEN");
        }
        ws.onclose = function(evt) {
            print("CLOSE");
            ws = null;
        }
        ws.onmessage = function(evt) {
            if(evt.data.match("http", evt.data)){
              showGif(evt.data);
            } else if(evt.data.match("keepalive", evt.data)){
              console.log("keepalive");
            } else {
              print("RESPONSE: " + evt.data);
            }
        }
        ws.onerror = function(evt) {
            print("ERROR: " + evt.data);
        }
        return false;
    };
    document.getElementById("send").onclick = function(evt) {
        if (!ws) {
            return false;
        }
        print("SEND: " + input.value);
        ws.send(input.value);
        return false;
    };
    document.getElementById("close").onclick = function(evt) {
        if (!ws) {
            return false;
        }
        ws.close();
        return false;
    };
});