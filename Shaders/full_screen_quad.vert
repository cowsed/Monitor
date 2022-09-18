#version 330 core

in vec3 pos;
out vec2 UV;

void main(){
    UV = (pos.xy+1)/2;
    gl_Position = vec4(pos, 1.0);
}