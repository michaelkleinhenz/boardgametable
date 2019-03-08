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
  console.log("setting colomap " + colormap + " and brightness to " + brightness);
  var xmlHttp = new XMLHttpRequest();
  xmlHttp.open("GET", "http://192.168.178.86:8080/api?command=startcolormap&map=" + colormap + "&brightness=" + brightness, false);
  xmlHttp.send(null);
  console.log("response: "+ xmlHttp.status);
}
function setBrightness(brightness) {
/* brightness can not be set individually yet
  console.log("setting brightness to " + brightness);
  var xmlHttp = new XMLHttpRequest();
  xmlHttp.open("GET", "http://127.0.0.1:8080/api?command=brightness&value="+ brightness, false);
  xmlHttp.send(null);
  console.log("Response: "+ xmlHttp.status);
*/
}
