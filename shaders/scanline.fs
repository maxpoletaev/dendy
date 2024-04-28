#version 330

// Input vertex attributes (from vertex shader)
in vec2 fragTexCoord;
in vec4 fragColor;

// Input uniform values
uniform sampler2D texture0;
uniform vec4 colDiffuse;

// Output fragment color
out vec4 finalColor;

uniform float time;

void main()
{
    // Fetch the base texel color
    vec4 baseColor = texture(texture0, fragTexCoord);

    // Dynamic Chromatic Aberration Effect
    float caIntensity = 0.001 * sin(time * 2.0) + 0.0015; // Modulates over time
    vec4 colorR = texture(texture0, fragTexCoord + vec2(caIntensity, 0.0));
    vec4 colorG = texture(texture0, fragTexCoord);
    vec4 colorB = texture(texture0, fragTexCoord - vec2(caIntensity, 0.0));
    vec3 chromaColor = vec3(colorR.r, colorG.g, colorB.b); // Chromatic aberration effect

    // Scanline parameters: Adjustable width and spacing
    float lineSpacing = 5.0; // Distance between the start of one scanline to the next
    float lineWidth = 2.0;  // Width of each scanline
    float yPos = mod(gl_FragCoord.y, lineSpacing);
    float scanlineEffect = (yPos < lineWidth) ? 0.90 : 1.0; // Apply darkening on the scanlines

    // Combine the effects
    vec3 mixedColor = mix(baseColor.rgb, chromaColor, 0.95); // Blend chromatic aberration and base color
    vec3 finalColorRGB = mixedColor * scanlineEffect; // Apply scanline effect with modulation

    // Output the final color with scanline effect and color tint
    finalColor = vec4(finalColorRGB, 1.0) * colDiffuse;
}
