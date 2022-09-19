#version 330 core

uniform sampler2D screenImage;

uniform ivec2 physical_border_dims = ivec2(8,8);
uniform ivec2 screen_dims = ivec2(512,512); // of terminal
uniform int ScanlinePosition = 0;
uniform float scanlineStrength = .13;
uniform float noiseStrength = .03;

uniform float ambient = .12;
uniform float text_brightness = .7;

in vec2 inUV;
out vec4 color;

vec2 flipY(vec2 iv){
    return vec2(iv.x, 1-iv.y);
}

float PHI = 1.61803398874989484820459;  // Φ = Golden Ratio   

float gold_noise(in vec2 xy, in float seed){
       return fract(tan(distance(xy*PHI, xy)*seed)*xy.x);
}


float sdRoundedBox( in vec2 p, in vec2 b, in vec4 r )
{
    r.xy = (p.x>0.0)?r.xy : r.zw;
    r.x  = (p.y>0.0)?r.x  : r.y;
    vec2 q = abs(p)-b+r.x;
    return min(max(q.x,q.y),0.0) + length(max(q,0.0)) - r.x;
}

void main(){


    vec2 UV = inUV;
    UV *=vec2(screen_dims+physical_border_dims*2)/vec2(screen_dims);
    UV -= vec2(physical_border_dims)/vec2(screen_dims);

    ivec2 pix = ivec2(flipY(UV)*screen_dims);



    //Scanline Addition
    int index = pix.y*screen_dims.x + pix.x;
    int numPix = screen_dims[0]*screen_dims[1];
    index-=ScanlinePosition;
    index+=numPix;
    index = index%(numPix+screen_dims[0]);
    float scanAddition = float(index)/float(numPix);
    scanAddition = pow(scanAddition, 2.0);
    //Noise
    float noiseAddition = gold_noise(vec2(pix), float(12+(ScanlinePosition/(2*512))%30000));
    noiseAddition = float(noiseAddition > .5);

    vec3 termCol = texture(screenImage, flipY(UV)).xyz;
    
    vec3 col = vec3(ambient) + termCol * text_brightness;
    
    //col += scanAddition * scanlineStrength ;
    col += ((noiseAddition) * noiseStrength);

    //col += .2 * mix(vec3(0), vec3(1), float((pix.x%2==0) != (pix.y%2==0)));

    float dist = sdRoundedBox((UV-.5), vec2(.3)
    , vec4(.2));
    float fast_dist = sdRoundedBox((UV-.5)*1.04, vec2(.5), vec4(.06));

    if (dist>0){
        col -= dist * .3;
    }

    //fast fade out at edges
    if (fast_dist>0){
        col -= pow(fast_dist*1.4,.8);
    }

    if (UV.x<0 || UV.x>1 || UV.y<0 || UV.y>1){
        col = vec3(0);
    }

    color = vec4(col,1);
}