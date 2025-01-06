const go = new Go();

const documentReady = new Promise((resolve) => {
  if (document.readyState !== "loading") {
    resolve();
  } else {
    document.addEventListener("DOMContentLoaded", resolve);
  }
});

const wasmReady = WebAssembly.instantiateStreaming(fetch("dendy.wasm"), go.importObject).then((result) => {
  go.run(result.instance);
});

Promise.all([wasmReady, documentReady]).then(() => {
  const width = 256;
  const height = 240;
  const targetFPS = 60;
  const scale = 2;

  let canvas = document.getElementById("canvas");
  canvas.width = width;
  canvas.height = height;
  canvas.style.width = width*scale + "px";
  canvas.style.height = height*scale + "px";
  canvas.style.imageRendering = "pixelated";

  let ctx = canvas.getContext("2d");
  ctx.imageSmoothingEnabled = false;

  let imageData = ctx.createImageData(width, height);
  let buttonsPressed = 0;

  const BUTTON_A = 1 << 0;
  const BUTTON_B = 1 << 1;
  const BUTTON_SELECT = 1 << 2;
  const BUTTON_START = 1 << 3;
  const BUTTON_UP = 1 << 4;
  const BUTTON_DOWN = 1 << 5;
  const BUTTON_LEFT = 1 << 6;
  const BUTTON_RIGHT = 1 << 7;

  const keyMap = {
    "KeyW": BUTTON_UP,
    "KeyS": BUTTON_DOWN,
    "KeyA": BUTTON_LEFT,
    "KeyD": BUTTON_RIGHT,
    "Enter": BUTTON_START,
    "ShiftRight": BUTTON_SELECT,
    "KeyJ": BUTTON_B,
    "KeyK": BUTTON_A,
  };

  document.addEventListener("keydown", (event) => {
    if (keyMap[event.code]) {
      event.preventDefault();
      buttonsPressed |= keyMap[event.code];
    }
  });

  document.addEventListener("keyup", (event) => {
    if (keyMap[event.code]) {
      event.preventDefault();
      buttonsPressed &= ~keyMap[event.code];
    }
  });

  document.getElementById("file-input").addEventListener("input", function (){
    this.files[0].arrayBuffer().then((buffer) => {
      let rom = new Uint8Array(buffer);
      uploadROM(rom);
    });
    this.blur();
  });

  setInterval(() => {
    runFrame(imageData.data, buttonsPressed);
    ctx.putImageData(imageData, 0, 0);
  }, 1000 / targetFPS);
});