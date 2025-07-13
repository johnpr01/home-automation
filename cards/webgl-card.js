import { LitElement, html, css } from "https://unpkg.com/lit-element@2.0.1/lit-element.js?module";
import * as THREE from "https://unpkg.com/three@0.128.0/build/three.module.js";

class WebGLCard extends LitElement {
  static get properties() {
    return {
      _hass: {},
      config: {},
    };
  }

  constructor() {
    super();
  }

  // Define the card's rendering
  render() {
    return html`
      <ha-card>
        <canvas id="webglCanvas"></canvas>
      </ha-card>
    `;
  }

  // Lifecycle callback when the card is first connected to the DOM
  firstUpdated() {
    const canvas = this.shadowRoot.getElementById("webglCanvas");

    // Basic Three.js setup
    const scene = new THREE.Scene();
    const camera = new THREE.PerspectiveCamera(75, canvas.clientWidth / canvas.clientHeight, 0.1, 1000);
    const renderer = new THREE.WebGLRenderer({ canvas });

    // Set renderer size
    renderer.setSize(canvas.clientWidth, canvas.clientHeight);

    // Create a cube
    const geometry = new THREE.BoxGeometry();
    const material = new THREE.MeshBasicMaterial({ color: 0x00ff00 });
    const cube = new THREE.Mesh(geometry, material);
    scene.add(cube);

    camera.position.z = 5;

    // Animation loop
    const animate = () => {
      requestAnimationFrame(animate);
      cube.rotation.x += 0.01;
      cube.rotation.y += 0.01;
      renderer.render(scene, camera);
    };

    animate();
  }

  // Lifecycle callback when Home Assistant state changes
  set hass(hass) {
    this._hass = hass;
    // You can access this._hass to update your WebGL scene based on Home Assistant state changes
  }

  // Configuration (optional, if your card has customizable properties)
  setConfig(config) {
    this.config = config;
  }

  // Define the size of the card for the dashboard layout
  getCardSize() {
    return 5; // A height of 5 is equivalent to 250 pixels
  }

  static get styles() {
    return css`
      ha-card {
        height: 250px;
      }
      canvas {
        width: 100%;
        height: 100%;
        display: block;
      }
    `;
  }
}

customElements.define("webgl-card", WebGLCard);
