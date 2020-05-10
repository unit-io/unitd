// Called after form input is processed
function startConnect() {
    // Fetch the hostname/IP address and port number from the form
    host = document.getElementById("host").value;
    port = document.getElementById("port").value;
    clientID = document.getElementById("clientid").value;

    // Print output for the user in the messages div
    document.getElementById("messages").innerHTML += '<span>Connecting to: ' + host + ' on port: ' + port + '</span><br/>';
    document.getElementById("messages").innerHTML += '<span>Using the following client value: ' + clientID + '</span><br/>';

    // Initialize new Paho client connection
    client = new Paho.MQTT.Client(host, Number(port), clientID);

    // Set callback handlers
    client.onConnectionLost = onConnectionLost;
    client.onMessageArrived = onMessageArrived;

    // Connect the client, if successful, call onConnect function
    client.connect({
        onSuccess: onConnect,
    });
}

// Called after form input is processed
function genClientId() {
    // Fetch the hostname/IP address and port number from the form
    host = document.getElementById("host").value;
    port = document.getElementById("port").value;
    clientID = document.getElementById("clientid").value;

    // Print output for the user in the messages div
    document.getElementById("messages").innerHTML += '<span>Connecting to: ' + host + ' on port: ' + port + '</span><br/>';
    document.getElementById("messages").innerHTML += '<span>Using the following client value: ' + clientID + '</span><br/>';

    // Initialize new Paho client connection
    client = new Paho.MQTT.Client(host, Number(port), clientID);

    // Set callback handlers
    client.onConnectionLost = onConnectionLost;
    client.onMessageArrived = onMessageArrived;

    // Connect the client, if successful, call onConnect function
    client.connect({
        onSuccess: onConnect,
    });
}

// Called after form input is processed
function startConnect() {
    // Fetch the hostname/IP address and port number from the form
    host = document.getElementById("host").value;
    port = document.getElementById("port").value;
    clientID = document.getElementById("clientid").value;

    // Print output for the user in the messages div
    document.getElementById("messages").innerHTML += '<span>Connecting to: ' + host + ' on port: ' + port + '</span><br/>';
    document.getElementById("messages").innerHTML += '<span>Using the following client value: ' + clientID + '</span><br/>';

    // Initialize new Paho client connection
    client = new Paho.MQTT.Client(host, Number(port), clientID);

    // Set callback handlers
    client.onConnectionLost = onConnectionLost;
    client.onMessageArrived = onMessageArrived;

    // Connect the client, if successful, call onConnect function
    client.connect({
        onSuccess: onConnect,
    });
}

// Called after form input is processed
function genKey() {
    // Fetch the MQTT topic from the form
    topic = document.getElementById("topic").value;
    to = document.getElementById("to").value;

    if (!topic.includes("/")) {
        payload = JSON.stringify({ "topic": topic, "type": "rw" });
        message = new Paho.MQTT.Message(payload);
        message.destinationName = "unitd/keygen";
        client.send(message);
    }
    if (!to.includes("/")) {
        payload = JSON.stringify({ "topic": to, "type": "rw" });
        message = new Paho.MQTT.Message(payload);
        message.destinationName = "unitd/keygen";
        client.send(message);
    }
}

// Called after form input is processed
function onSubscribe() {
    // Fetch the MQTT topic from the form
    topic = document.getElementById("topic").value;

    // Print output for the user in the messages div
    document.getElementById("messages").innerHTML += '<span>Subscribing to: ' + topic + '</span><br/>';

    // Subscribe to the requested topic
    client.subscribe(topic);
}

// Called after form input is processed
function onPublish() {
    // Fetch the MQTT topic from the form
    to = document.getElementById("to").value;
    msg = document.getElementById("msg").value;

    message = new Paho.MQTT.Message(msg);
    message.destinationName = to + "?ttl=3m";
    client.send(message);
}

// Called when the client connects
function onConnect() {
    // Fetch the MQTT topic from the form
    topic = document.getElementById("topic").value;
    to = document.getElementById("to").value;
    msg = document.getElementById("msg").value;

    if (topic.includes("/")) {
        // Print output for the user in the messages div
        document.getElementById("messages").innerHTML += '<span>Subscribing to: ' + topic + '</span><br/>';

        // Subscribe to the requested topic
        client.subscribe(topic);
    }
    if (to.includes("/")) {
        // Public message to topic
        message = new Paho.MQTT.Message(msg);
        message.destinationName = to + "?ttl=3m";
        client.send(message);
    }
}

// Called when the client loses its connection
function onConnectionLost(responseObject) {
    document.getElementById("messages").innerHTML += '<span>ERROR: Connection lost</span><br/>';
    if (responseObject.errorCode !== 0) {
        document.getElementById("messages").innerHTML += '<span>ERROR: ' + + responseObject.errorMessage + '</span><br/>';
    }
}

// Called when a message arrives
function onMessageArrived(message) {
    console.log("onMessageArrived: " + message.payloadString);
    document.getElementById("messages").innerHTML += '<span>' + message.payloadString + '</span><br/>';
    updateScroll();
}

// Called when the disconnection button is pressed
function startDisconnect() {
    client.disconnect();
    document.getElementById("messages").innerHTML += '<span>Disconnected</span><br/>';
}

// Updates #messages div to auto-scroll
function updateScroll() {
    var element = document.getElementById("messages");
    element.scrollTop = element.scrollHeight;
}