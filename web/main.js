const documentReady = new Promise((resolve) => {
  if (document.readyState !== "loading") {
    resolve();
  } else {
    document.addEventListener("DOMContentLoaded", resolve);
  }
});

window.go = new Go();
const wasmReady = WebAssembly.instantiateStreaming(fetch("dendy.wasm"), go.importObject).then((result) => {
  go.run(result.instance);
});

Promise.all([wasmReady, documentReady]).then(async () => {
  const WIDTH = 256;
  const HEIGHT = 240;
  const TARGET_FPS = 60;

  // ========================
  // Canvas setup
  // ========================

  let canvas = document.getElementById("canvas");
  canvas.width = WIDTH;
  canvas.height = HEIGHT;
  canvas.style.imageRendering = "pixelated";

  let ctx = canvas.getContext("2d");
  ctx.imageSmoothingEnabled = false;

  // ========================
  //  Audio setup
  // ========================

  const audioBufferSize = go.AudioBufferSize;
  const audioSampleRate = go.AudioSampleRate;
  console.log(`[INFO] audio sample rate: ${audioSampleRate}, buffer size: ${audioBufferSize}`);
  let audioCtx = new AudioContext({sampleRate: audioSampleRate});

  await audioCtx.audioWorklet.addModule("audio.js");
  let audioNode = new AudioWorkletNode(audioCtx, "audio-processor");
  audioNode.connect(audioCtx.destination);

  // ========================
  //  Mute/unmute button
  // ========================

  let unmuteButton = document.getElementById("unmute-button");
  if (audioCtx.state === "suspended") {
    unmuteButton.style.display = "block";
  }

  document.addEventListener("click", function() {
    if (audioCtx.state === "suspended") {
      unmuteButton.style.display = "none";
      audioCtx.resume();
    }
  }, {once: true});

  // ========================
  //  Input handling
  // ========================

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
    "LeftShift": BUTTON_SELECT,
    "KeyJ": BUTTON_B,
    "KeyK": BUTTON_A,
  };

  let buttonsPressed = 0;

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

  const elementKeyMap = {
    "dpad-up": BUTTON_UP,
    "dpad-down": BUTTON_DOWN,
    "dpad-left": BUTTON_LEFT,
    "dpad-right": BUTTON_RIGHT,
    "button-start": BUTTON_START,
    "button-select": BUTTON_SELECT,
    "button-b": BUTTON_B,
    "button-a": BUTTON_A,
  };

  for (let [id, mask] of Object.entries(elementKeyMap)) {
    let el = document.getElementById(id);

    el.addEventListener("mousedown", (e) => {
      e.preventDefault();
      buttonsPressed |= mask;
    });

    el.addEventListener("touchstart", (e) => {
      e.preventDefault();
      buttonsPressed |= mask;
    });

    el.addEventListener("mouseup", (e) => {
      e.preventDefault();
      buttonsPressed &= ~mask;
    });

    el.addEventListener("touchend", (e) => {
      e.preventDefault();
      buttonsPressed &= ~mask;
    });
  }

  document.querySelectorAll(".controls *").forEach((el) => {
    el.style.touchAction = "manipulation";
  });

  // ========================
  //  ROM loading
  // ========================

  let fileInput = document.getElementById("file-input");

  fileInput.addEventListener("input", function () {
    this.files[0].arrayBuffer().then((buffer) => {
      let rom = new Uint8Array(buffer);
      let ok = go.LoadROM(rom);
      if (!ok) {
        alert("Invalid or unsupported ROM file");
        this.value = "";
      }
    });
    this.blur(); // Avoid re-opening file dialog when pressing Enter
  });

  if (fileInput.files.length > 0) {
    fileInput.files[0].arrayBuffer().then((buffer) => {
      let rom = new Uint8Array(buffer);
      let ok = go.LoadROM(rom);
      if (!ok) {
        fileInput.value = "";
      }
    });
  }

  document.addEventListener("dragover", (e) => {
    e.preventDefault();
  });

  document.addEventListener("drop", (e) => {
    e.preventDefault();
    fileInput.files = e.dataTransfer.files;
    fileInput.dispatchEvent(new Event("input"));
  });

  // ========================
  //  Game loop
  // ========================

  function isInFocus() {
    return document.hasFocus() && document.visibilityState === "visible";
  }

  function getMemoryBuffer() {
    return go._inst.exports.mem?.buffer || go._inst.exports.memory.buffer; // latter is for TinyGo
  }

  function executeFrame() {
    while (true) {
      let frameReady = go.RunFrame(buttonsPressed);

      if (frameReady) {
        let framePtr = go.GetFrameBufferPtr();
        let image = new ImageData(new Uint8ClampedArray(getMemoryBuffer(), framePtr, WIDTH * HEIGHT * 4), WIDTH, HEIGHT);
        ctx.putImageData(image, 0, 0);
        return;
      }

      let audioBufPtr = go.GetAudioBufferPtr();
      let audioBuf = new Float32Array(getMemoryBuffer(), audioBufPtr, go.AudioBufferSize);
      audioNode.port.postMessage(audioBuf.slice()); // TODO: avoid copy
    }
  }

  let lastFrameTime = performance.now();
  const frameTime = 1000 / TARGET_FPS;

  function loop() {
    requestAnimationFrame(loop);

    const now = performance.now();
    const elapsed = now - lastFrameTime;
    if (elapsed < frameTime) return;

    const excessTime = elapsed % frameTime;
    lastFrameTime = now - excessTime;

    if (isInFocus()) {
      executeFrame();
    }
  }

  requestAnimationFrame(loop);
});
