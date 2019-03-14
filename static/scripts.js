function setColor(direction, color) {
  var arr = color.split('');
  var leds = "0,40,";
  if (direction==1)
    leds = "45,115,";
  else if (direction==2)
    leds = "120,156,";
  else if (direction==3)
    leds = "165,236,";
  var colormap = leds+arr[1]+arr[2]+","+arr[3]+arr[4]+","+arr[5]+arr[6];
  var brightness = document.getElementById("brightness").value;
  console.log("setting colormap " + colormap + " and brightness to " + brightness);
  var xmlHttp = new XMLHttpRequest();
  xmlHttp.open("GET", "/api?command=startcolormap&map=" + colormap + "&brightness=" + brightness, false);
  xmlHttp.send(null);
  console.log("response: "+ xmlHttp.status);
}

function updateColors() {
  var colorLeft = document.getElementById("colorLeft").value.split('');
  var colorRight = document.getElementById("colorRight").value.split('');
  var colorTop = document.getElementById("colorTop").value.split('');
  var colorBottom = document.getElementById("colorBottom").value.split('');
  var colormapLeft = "120,156,"+colorLeft[1]+colorLeft[2]+","+colorLeft[3]+colorLeft[4]+","+colorLeft[5]+colorLeft[6];
  var colormapRight = "0,40,"+colorRight[1]+colorRight[2]+","+colorRight[3]+colorRight[4]+","+colorRight[5]+colorRight[6];
  var colormapTop = "165,236,"+colorTop[1]+colorTop[2]+","+colorTop[3]+colorTop[4]+","+colorTop[5]+colorTop[6];
  var colormapBottom = "45,115,"+colorBottom[1]+colorBottom[2]+","+colorBottom[3]+colorBottom[4]+","+colorBottom[5]+colorBottom[6];
  var colormap = colormapLeft+"-"+colormapRight+"-"+colormapTop+"-"+colormapBottom;
  var brightness = document.getElementById("brightness").value;
  console.log("setting all colors: " + colormap + " and brightness to " + brightness);
  var xmlHttp = new XMLHttpRequest();
  xmlHttp.open("GET", "/api?command=startcolormap&map=" + colormap + "&brightness=" + brightness, false);
  xmlHttp.send(null);
  console.log("response: "+ xmlHttp.status);
}

function setBrightness(brightness) {
  console.log("setting brightness to " + brightness);
  var xmlHttp = new XMLHttpRequest();
  xmlHttp.open("GET", "/api?command=brightness&value="+ brightness, false);
  xmlHttp.send(null);
  console.log("Response: "+ xmlHttp.status);
}

function setActive(direction) {
  console.log("setting direction active: " + direction);
  var xmlHttp = new XMLHttpRequest();
  xmlHttp.open("GET", "/api?command=active&direction="+ direction, false);
  xmlHttp.send(null);
  console.log("Response: "+ xmlHttp.status);
}

function disableActive() {
  console.log("disabling direction active");
  var xmlHttp = new XMLHttpRequest();
  xmlHttp.open("GET", "/api?command=activeoff", false);
  xmlHttp.send(null);
  console.log("Response: "+ xmlHttp.status);
}

function reconnect() {
  console.log("reconnect controller");
  var xmlHttp = new XMLHttpRequest();
  xmlHttp.open("GET", "/api?command=reconnect", false);
  xmlHttp.send(null);
  console.log("Response: "+ xmlHttp.status);
}