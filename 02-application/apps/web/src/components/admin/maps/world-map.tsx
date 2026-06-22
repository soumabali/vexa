'use client';

import React from 'react';
import { ComposableMap, Geographies, Geography, Marker, ZoomableGroup } from 'react-simple-maps';

interface MapPoint {
  lat: number;
  lng: number;
  country: string;
  connections: number;
}

interface WorldMapProps {
  points: MapPoint[];
}

const geoUrl = 'https://cdn.jsdelivr.net/npm/world-atlas@2/countries-110m.json';

export default function WorldMap({ points }: WorldMapProps) {
  const maxConnections = Math.max(...points.map((p) => p.connections));

  return (
    <ComposableMap projection="geoMercator" projectionConfig={{ scale: 100 }}>
      <ZoomableGroup center={[0, 20]} zoom={1}>
        <Geographies geography={geoUrl}>
          {({ geographies }) =>
            geographies.map((geo) => (
              <Geography
                key={geo.rsmKey}
                geography={geo}
                fill="#e5e7eb"
                stroke="#d1d5db"
                strokeWidth={0.5}
                style={{
                  default: { outline: 'none' },
                  hover: { fill: '#d1d5db', outline: 'none' },
                  pressed: { fill: '#9ca3af', outline: 'none' },
                }}
              />
            ))
          }
        </Geographies>
        {points.map((point, index) => {
          const size = Math.max(4, (point.connections / maxConnections) * 20);
          return (
            <Marker key={index} coordinates={[point.lng, point.lat]}>
              <circle
                r={size}
                fill="#10b981"
                opacity={0.6}
                stroke="#fff"
                strokeWidth={1}
              />
              <text
                textAnchor="middle"
                y={size + 12}
                style={{ fontSize: '10px', fill: '#374151' }}
              >
                {point.country}: {point.connections}
              </text>
            </Marker>
          );
        })}
      </ZoomableGroup>
    </ComposableMap>
  );
}