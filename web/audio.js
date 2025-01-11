class AudioProcessor extends AudioWorkletProcessor {
  constructor() {
    super();
    this.position = 0;
    this.buffers = [];
    this.current = null;

    this.port.onmessage = (e) => {
      this.buffers.push(e.data);
      if (this.buffers.length > 3) {
        this.buffers.shift();
      }
    };
  }

  process(inputs, outputs, parameters) {
    let output = outputs[0];
    let channel = output[0];

    if (!this.current || this.position >= this.current.length) {
      this.current = this.buffers.shift();
      this.position = 0;

      if (!this.current) {
        for (let i = 0; i < channel.length; i++) {
          channel[i] = 0;
        }
        return true;
      }
    }

    for (let i = 0; i < channel.length; i++) {
      channel[i] = this.current[this.position++];
    }

    return true;
  }
}

registerProcessor("audio-processor", AudioProcessor);