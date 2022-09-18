#version 330 core

uniform sampler2D screenImage;


in vec2 UV;
out vec4 color;

vec2 flipY(vec2 iv){
    return vec2(iv.x, 1-iv.y);
}

void main(){
    
    vec3 col = texture(screenImage, flipY(UV)).xyz;

    color = vec4(col,1);
}