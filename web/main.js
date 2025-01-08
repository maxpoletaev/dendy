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

const peerReady = new Promise((resolve) => {
  let params = new URLSearchParams(location.search);
  let peer = new Peer(params.get("peer"), {
    debug: 2,
  });
  peer.on("open", (id) => {
    console.log("my peer id:", id);
    resolve(peer);
  });
});

Promise.all([wasmReady, documentReady, peerReady]).then(async (values) => {
  let peer = values[2];

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

  document.addEventListener("keydown", (e) => {
    if (keyMap[e.code]) {
      e.preventDefault();
      buttonsPressed |= keyMap[e.code];
    }
  });

  document.addEventListener("keyup", (e) => {
    if (keyMap[e.code]) {
      e.preventDefault();
      buttonsPressed &= ~keyMap[e.code];
    }
  });

  function startGameLoop(channel, isHost) {
    startSession(channel, isHost);

    let nextFrame = () => {
      let start = performance.now();
      runFrame(imageData.data, buttonsPressed);
      // runFrame(null, buttonsPressed);
      ctx.putImageData(imageData, 0, 0);
      let elapsed = performance.now() - start;
      let nextTimeout = Math.max(0, (1000/targetFPS) -elapsed);
      setTimeout(nextFrame, nextTimeout);
    };

    nextFrame();
  }

  async function connectToPeer(peerId) {
    return new Promise((resolve) => {
      let conn = peer.connect(peerId, {
        serialization: "raw",
        reliable: true,
      });
      conn.on("open", () => {
        channel = conn.dataChannel;
        resolve(channel);
      });
    });
  }

  async function waitForPeerConnected() {
    return new Promise((resolve) => {
      peer.on("connection", (conn) => {
        conn.on("open", () => {
          channel = conn.dataChannel;
          resolve(channel);
        });
      });
    });
  }

  async function waitFileSelected() {
    return new Promise((resolve) => {
      document.getElementById("file-input").addEventListener("input", function() {
        this.files[0].arrayBuffer().then((buffer) => {
          let rom = new Uint8Array(buffer);
          uploadROM(rom);
          resolve();
        });
        this.blur();
      });

      if (document.getElementById("file-input").files.length) {
        document.getElementById("file-input").dispatchEvent(new Event("input"));
      }
    });
  }

  let params = new URLSearchParams(location.search);
  let peerId = params.get("target");
  let remoteReady = false;
  let channel = null;
  let isHost = false;

  if (peerId) {
    console.log("connecting to peer:", peerId);
    channel = await connectToPeer(peerId);
    console.log("connected")
  } else {
    console.log("waiting for peer to connect");
    channel = await waitForPeerConnected();
    console.log("peer connected");
    isHost = true;
  }

  channel.addEventListener("message", (e) => {
    if (e.data === "ready") {
      console.log("remote side is ready");
      remoteReady = true;
    }
  });

  console.log("awaiting ROM");
  await waitFileSelected();
  console.log("local side is ready");
  channel.send("ready");

  let checkRemoteReady = () => {
    if (!remoteReady) {
      console.log("remote side is not ready yet");
      setTimeout(checkRemoteReady, 100);
      return;
    }
    console.log("starting game loop");
    setTimeout(() => startGameLoop(channel, isHost), 1000);
  };

  checkRemoteReady();
});