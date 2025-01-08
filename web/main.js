const documentReady = new Promise((resolve) => {
  if (document.readyState !== "loading") {
    resolve();
  } else {
    document.addEventListener("DOMContentLoaded", resolve);
  }
});

const go = new Go();
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

  let fileInput = document.getElementById("file-input");

  fileInput.addEventListener("input", function() {
    this.files[0].arrayBuffer().then((buffer) => {
      let rom = new Uint8Array(buffer);
      let ok = uploadROM(rom);
      if (!ok) {
        alert("Invalid ROM file");
        this.value = "";
      }
    });
    this.blur();
  });

  if (fileInput.files.length > 0) {
    fileInput.files[0].arrayBuffer().then((buffer) => {
      let rom = new Uint8Array(buffer);
      let ok = uploadROM(rom);
      if (!ok) {
        fileInput.value = "";
      }
    });
  }

  function isInFocus() {
    return document.hasFocus() && document.visibilityState === "visible";
  }

  function gameLoop() {
    let nextFrame = () => {
      let start = performance.now();

      if (isInFocus()) {
        let framePtr = runFrame(buttonsPressed);
        let memPtr = go._inst.exports.mem?.buffer || go._inst.exports.memory.buffer; // latter is for TinyGo
        let image = new ImageData(new Uint8ClampedArray(memPtr, framePtr, width * height * 4), width, height);
        ctx.putImageData(image, 0, 0);
      }

      let elapsed = performance.now() - start;
      let nextTimeout = Math.max(0, (1000 / targetFPS) - elapsed);
      setTimeout(nextFrame, nextTimeout);
    };

    nextFrame();
  }

  gameLoop();
});
