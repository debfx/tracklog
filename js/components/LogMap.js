"use strict";

import React from "react";
import ReactDOM from "react-dom";
import Leaflet from "leaflet";

const COLORS = ["red", "blue", "purple", "teal", "maroon"];

export default class LogMap extends React.Component {
  componentDidMount() {
    this.map = Leaflet.map(ReactDOM.findDOMNode(this));

    let layer;
    if (this.props.mapboxAccessToken) {
      layer = Leaflet.tileLayer(`https://api.mapbox.com/styles/v1/mapbox/outdoors-v11/tiles/256/{z}/{x}/{y}?access_token=${this.props.mapboxAccessToken}`, {
        attribution: `<a href="https://www.mapbox.com/about/maps/" target="_blank">Terms &amp; Feedback</a>`,
      });
    } else {
      layer = Leaflet.tileLayer("https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png");
    }

    this.polylines = [];
    this.map.addLayer(layer);
    this.updateMap();
  }

  componentWillUnmount() {
    this.map.remove();
    this.map = null;
  }

  updateMap() {
    for (const polyline of this.polylines) {
      this.map.removeLayer(polyline);
    }
    this.polylines = [];

    const latlngs = this.props.log.get("tracks").map((track) => {
      return track.map((point) => {
        return [point.get("lat"), point.get("lon")];
      });
    }).toJS();

    for (let i = 0; i < latlngs.length; i++) {
        let polyline = Leaflet.polyline(latlngs[i], { color: COLORS[i % COLORS.length], weight: 3, opacity: 0.8 });
        this.polylines.push(polyline);
        polyline.addTo(this.map);
    }
    let multiPolyline = Leaflet.polyline(latlngs);
    this.map.fitBounds(multiPolyline.getBounds());
  }

  render() {
    if (this.map) {
      this.updateMap();
    }

    return <div className="log-map"></div>;
  }
}
