import React, {useEffect, useState} from 'react';
import DeckGL from '@deck.gl/react';
import {MVTLayer} from '@deck.gl/geo-layers';
import {StaticMap} from "react-map-gl";

const TRIP_URL = 'http://localhost:7800/public.trip_last_instant/{z}/{x}/{y}.pbf';
const LINE_URL = 'http://localhost:7800/public.line_last_instant/{z}/{x}/{y}.pbf';
const TRAJECTORY_URL = 'http://localhost:7800/public.default_trajectory/{z}/{x}/{y}.pbf';
const MAP_STYLE = 'https://basemaps.cartocdn.com/gl/dark-matter-nolabels-gl-style/style.json';

const INITIAL_VIEW_STATE = {
    longitude: 4.383406,
    latitude: 50.815338,
    zoom: 11,
    minZoom: 0,
    maxZoom: 23,
};

function App() {

    const [tripParameter, setTripParameter] = useState('');
    const [lineParameter, setLineParameter] = useState('');
    const [trajectoryParameter, setTrajectoryParameter] = useState('');
    const [tripURL, setTripURL] = useState('');
    const [lineURL, setLineURL] = useState('');
    const [trajectoryURL, setTrajectoryURL] = useState('');
    const [tripLayerVisibility, setTripLayerVisibility] = useState(false);
    const [lineLayerVisibility, setLineLayerVisibility] = useState(false);
    const [trajectoryLayerVisibility, setTrajectoryLayerVisibility] = useState(false);

    const onTileLoad = (tile) => {
        console.log(tile.content);
    }

    const layers = [
        new MVTLayer({
            id: 'tripLayer',
            data: tripURL,
            onTileLoad : onTileLoad ,
            minZoom: 0,
            maxZoom: 22,
            getPointRadius: 7,
            getFillColor: [255, 0, 0],
            visible: tripLayerVisibility,
        }),
        new MVTLayer({
            id: 'lineLayer',
            data: lineURL,
            minZoom: 0,
            maxZoom: 22,
            getPointRadius: 7,
            getFillColor: [0, 0, 255],
            visible: lineLayerVisibility
        }),
        new MVTLayer({
            id: 'trajectoryLayer',
            data: trajectoryURL,
            minZoom: 0,
            maxZoom: 22,
            getLineColor: (d) => {
                let id = d.properties.shape_id;
                if (id.includes('b')){
                    return [255,0,0]
                } else if (id.includes('m')) {
                    return [0,255,255]
                } else if (id.includes('t')){
                    return [255,255,0]
                }
            },
            getLineWidth : 2,
            visible: trajectoryLayerVisibility,
        })
    ];

    // Handle input changes and button clicks
    const handleTripInputChange = (e) => setTripParameter(e.target.value);
    const handleLineInputChange = (e) => setLineParameter(e.target.value);
    const handleTrajectoryInputChange = (e) => setTrajectoryParameter(e.target.value);
    const toggleTripLayerVisibility = () => setTripLayerVisibility(!tripLayerVisibility);
    const toggleLineLayerVisibility = () => setLineLayerVisibility(!lineLayerVisibility);
    const toggleTrajectoryLayerVisibility = () => setTrajectoryLayerVisibility(!trajectoryLayerVisibility);
    const updateTripParameter = () => {
        console.log(tripParameter);
        setTripURL(TRIP_URL + `?p_tripid=` + tripParameter + `&timestamp=${new Date().getTime()}`);
    }
    const updateLineParameter = () => {
        console.log(lineParameter);
        setLineURL(LINE_URL + `?p_lineid=` + lineParameter + `&timestamp=${new Date().getTime()}`);
    }
    const updateTrajectoryParameter = () => setTrajectoryURL(TRAJECTORY_URL + `?p_lineid=` + trajectoryParameter);


    useEffect(() => {
        const interval = setInterval(() => {
            console.log(tripParameter,lineParameter)
            console.log(tripURL,lineURL)
            if (tripLayerVisibility) {
                // Modify the URL to trigger a refresh
                setTripURL(TRIP_URL + `?p_tripid=` + tripParameter + `&timestamp=${new Date().getTime()}`);
            }
            if (lineLayerVisibility){
                setLineURL(LINE_URL + `?p_lineid=` + lineParameter + `&timestamp=${new Date().getTime()}`);
            }
        }, 5000);

        return () => clearInterval(interval);
    });

    return (
        <div style={{ position: 'relative', height: '100vh' }}>
            <DeckGL
                layers={layers}
                initialViewState={INITIAL_VIEW_STATE}
                controller={true}
            >
                <StaticMap mapStyle={MAP_STYLE} />
            </DeckGL>
            <div style={{ position: 'absolute', top: 0, left: 0, padding: '10px' }}>
                <div>
                    <input type="text" value={tripParameter} onChange={handleTripInputChange} />
                    <button onClick={updateTripParameter}>Update TripID</button>
                    <button onClick={toggleTripLayerVisibility}>
                        {tripLayerVisibility ? 'Hide TripID Layer' : 'Show TripID Layer'}
                    </button>
                </div>
                <div style={{ marginTop: '10px' }}>
                    <input type="text" value={lineParameter} onChange={handleLineInputChange} />
                    <button onClick={updateLineParameter}>Update LineID</button>
                    <button onClick={toggleLineLayerVisibility}>
                        {lineLayerVisibility ? 'Hide LineID Layer' : 'Show LineID Layer'}
                    </button>
                </div>
                <div style={{ marginTop: '10px' }}>
                    <input type="text" value={trajectoryParameter} onChange={handleTrajectoryInputChange} />
                    <button onClick={updateTrajectoryParameter}>Update LineID</button>
                    <button onClick={toggleTrajectoryLayerVisibility}>
                        {trajectoryLayerVisibility ? 'Hide trajectory Layer' : 'Show trajectory Layer'}
                    </button>
                </div>
            </div>
        </div>
    );
}

export default App;
