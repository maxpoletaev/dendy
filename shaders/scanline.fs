#version 330

// Input vertex attributes (from vertex shader)
in vec2 fragTexCoord;
in vec4 fragColor;

// Input uniform values
uniform sampler2D texture0;
uniform vec4 colDiffuse;
uniform float time;
uniform float scale;

// Output fragment color
out vec4 finalColor;

void main()
{
    // Fetch the base texel color
    vec4 baseColor = texture(texture0, fragTexCoord);

    // Chromatic Aberration Effect
    float caIntensity = 0.002;
    vec4 colorR = texture(texture0, fragTexCoord + vec2(caIntensity, 0.0));
    vec4 colorG = texture(texture0, fragTexCoord);
    vec4 colorB = texture(texture0, fragTexCoord - vec2(caIntensity, 0.0));
    vec3 chromaColor = vec3(colorR.r, colorG.g, colorB.b);

    // Scanline Effect
    float lineSpacing = 2.0 * scale; // Distance between the start of one scanline to the next
    float lineWidth = 2.0; // Width of each scanline
    float yPos = mod(gl_FragCoord.y, lineSpacing);
    float scanlineEffect = (yPos < lineWidth) ? 0.80 : 1.0;

    // Combine the effects
    vec3 mixedColor = mix(baseColor.rgb, chromaColor, 0.95);
    vec3 finalColorRGB = mixedColor * scanlineEffect;

    // Output the final color with scanline effect and color tint
    finalColor = vec4(finalColorRGB, 1.0) * colDiffuse;
}
