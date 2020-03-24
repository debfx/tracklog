"use strict";

import React from "react";
import ReactDOM from "react-dom";
import Leaflet from "leaflet";

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

    this.map.addLayer(layer);
    this.updateMap();
  }

  componentWillUnmount() {
    this.map.remove();
    this.map = null;
  }

  updateMap() {
    if (this.multiPolyline) {
      this.map.removeLayer(this.multiPolyline);
    }

    const latlngs = this.props.log.get("tracks").map((track) => {
      return track.map((point) => {
        return [point.get("lat"), point.get("lon")];
      });
    }).toJS();

    this.multiPolyline = Leaflet.polyline(latlngs, { color: "red", weight: 3, opacity: 0.8 });
    this.multiPolyline.addTo(this.map);
    this.map.fitBounds(this.multiPolyline.getBounds());
  }

  render() {
    if (this.map) {
      this.updateMap();
    }

    return <div className="log-map"></div>;
  }
}
