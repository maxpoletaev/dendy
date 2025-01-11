class AudioProcessor extends AudioWorkletProcessor {
  constructor() {
    super();
    this.position = 0;
    this.buffers = [];
    this.current = null;

    this.port.onmessage = (e) => {
      this.buffers.push(e.data);
      if (this.buffers.length > 2) {
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
        channel.fill(0);
        return true;
      }
    }

    channel.set(this.current.subarray(this.position, this.position + channel.length));
    this.position += channel.length;
    return true;
  }
}

registerProcessor("audio-processor", AudioProcessor);