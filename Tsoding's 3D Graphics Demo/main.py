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
    # stop: list[int] 
    
@dataclass
class VectorPair:
    start: int
    stop: int

def normal(v: Vector): # Coord Translation for orthogonal projection from top-left (0,0) to a center cordinates (normalization)
    newX = (v.x + 1)/2 * screenWidth
    newY = (1 - (v.y + 1)/2) * screenHeight
    return replace(v, x=newX, y=newY) # only works in a 1:1 ratio so how do we make it scale based on screen ratios?
    
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

def VectorPipeline(v: Vector, angle: float, dz: float):
    return normal(project(translate_z(rotate_xz(v, angle), dz)))

VectorPoints = [
    Vector(0.25, 0.25, 0.25), 
    Vector(-0.25, 0.25, 0.25),
    Vector(0.25, -0.25, 0.25),
    Vector(-0.25, -0.25, 0.25),
    
    Vector(0.25, 0.25, -0.25),
    Vector(-0.25, 0.25, -0.25),
    Vector(0.25, -0.25, -0.25),
    Vector(-0.25, -0.25, -0.25),
]

VectorPairs = [
    VectorPair(0, 1),  # Front face
    VectorPair(1, 3),
    VectorPair(3, 2),
    VectorPair(2, 0),
    VectorPair(4, 5),  # Back face
    VectorPair(5, 7),
    VectorPair(7, 6),
    VectorPair(6, 4),
    VectorPair(0, 4),  # Connecting edges
    VectorPair(1, 5),
    VectorPair(2, 6),
    VectorPair(3, 7),
]
            
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
    for vp in VectorPairs:
        line(VectorPipeline(VectorPoints[vp.start], angle, dz), VectorPipeline(VectorPoints[vp.stop], angle, dz))
        # point(normal(project(translate_z(rotate_xz(v, angle), dz))))
   

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

