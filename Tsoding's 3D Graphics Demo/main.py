# Orthogonal Projection
# (x, y, z)
# x' = x/z
# y' = y/z

import pygame
import math
from dataclasses import dataclass, replace

@dataclass
class Vector:
    x: float
    y: float
    z: float
    size: int = 10
    color: str = "green"

def normal(v: Vector): # Coord Translation for orthogonal projection from top-left (0,0) to a center cordinates (normalization)
    newX = (v.x + 1)/2 * screenWidth
    newY = (1 - (v.y + 1)/2) * screenHeight
    return replace(v, x=newX, y=newY)
    
def point(v: Vector):
    return pygame.draw.rect(screen, v.color, pygame.Rect(v.x - v.size/2, v.y - v.size/2, v.size, v.size))

def line(v1: Vector, v2: Vector):
    start = [v1.x, v1.y]
    end = [v2.x, v2.y]
    return pygame.draw.line(screen, v1.color, start, end, width=5)

def project(v: Vector): 
    xPrime = v.x / v.z
    yPrime = v.y / v.z
    return replace(v, x = xPrime, y = yPrime)

def translate_z(v: Vector, dz):
    return replace(v, z = v.z + dz)

def rotate_xz(v: Vector, angle: float):
    c = math.cos(angle)
    s = math.sin(angle)
    x = v.x * c - v.z * s
    z = v.x * s + v.z * c
    return replace(v, x = x, z= z)

VectorPoints = [
    Vector(0.25, 0.25, 0.25, color="red"),
    Vector(-0.25, 0.25, 0.25, color="blue"),
    Vector(0.25, -0.25, 0.25, color="yellow"),
    Vector(-0.25, -0.25, 0.25),
    
    Vector(0.25, 0.25, -0.25, color="red"),
    Vector(-0.25, 0.25, -0.25, color="blue"),
    Vector(0.25, -0.25, -0.25, color="yellow"),
    Vector(-0.25, -0.25, -0.25),
]

def VectorPipeline(v: Vector, angle: float, dz: float):
    return normal(project(translate_z(rotate_xz(v, angle), dz)))

            
defaultColor = "green"
pygame.init()
screenWidth = 900
screenHeight = 900
screen = pygame.display.set_mode((screenWidth, screenHeight))
clock = pygame.time.Clock()
renderFPS = 60
running = True
dt = 1/renderFPS
dz = 1
angle = 0

def anim():
    global dt, dz, angle
    # dz += 1 * dt
    angle += 0.5*math.pi*dt
    screen.fill("black")
    for v in VectorPoints:
        point(normal(project(translate_z(rotate_xz(v, angle), dz))))

    line(VectorPipeline(VectorPoints[0], angle, dz), VectorPipeline(VectorPoints[4], angle, dz))

while running:
    for event in pygame.event.get():
        if event.type == pygame.QUIT:
            running = False

    # fill the screen with a color to wipe away anything from last frame
    screen.fill("black")
    
    anim()

    # flip() the display to put your work on screen
    pygame.display.flip()

    clock.tick(renderFPS)  # limits FPS to 60

pygame.quit()

