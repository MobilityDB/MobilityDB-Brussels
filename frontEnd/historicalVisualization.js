import React, {useEffect, useState} from 'react';
import {render} from 'react-dom';
import DeckGL from '@deck.gl/react';
import {MVTLayer, TripsLayer} from '@deck.gl/geo-layers';
import {StaticMap} from "react-map-gl";
import {BrowserRouter as Router} from "react-router-dom";
import TimeSlider from "./Slider";


const MAP_STYLE = 'https://basemaps.cartocdn.com/gl/dark-matter-nolabels-gl-style/style.json';


const INITIAL_VIEW_STATE = {
    longitude: 4.383406,
    latitude: 50.815338,
    zoom: 11,
    minZoom: 0,
    maxZoom: 23,
};


// animation start and end in unix timestamp


export default function App({
                                trailLength = 180,
                                animationSpeed = 5,
                                DATA_TRIPS_URL =
                                    'http://localhost:7800/public.historical_trip/{z}/{x}/{y}.pbf?p_tripid=',
                                DATA_LINES_URL =
                                    'http://localhost:7800/public.historical_line/{z}/{x}/{y}.pbf?p_lineid='
                            }) {


    const [time, setTime] = useState(0);
    const [animation] = useState({});
    const [isAnimated,setIsAnimated] = useState(false)
    const [dataTrips,setDataTrips] = useState(DATA_TRIPS_URL);
    const [tripParameter, setTripParameter] = useState('');
    const [dataLines,setDataLines] = useState(DATA_LINES_URL);
    const [lineParameter, setLineParameter] = useState('');
    const [minTimestamp, setMinTimestamp] = useState(0);
    const [maxTimestamp, setMaxTimestamp] = useState(0);
    const [minTimeURL, setMinTimeURL] = useState('');
    const [maxTimeURL, setMaxTimeURL] = useState('');
    const [loopLength,setLoopLength] = useState(1)

    const animate = () => {
        if (isAnimated){
            setTime(t => (t + animationSpeed) % loopLength);
            animation.id = window.requestAnimationFrame(animate);
        }
    };

    useEffect(
        () => {
            console.log(isAnimated);
            if  (isAnimated){
                animation.id = window.requestAnimationFrame(animate);
            }
            return () => window.cancelAnimationFrame(animation.id);
        },
        [animation,isAnimated]
    );

    const onTileLoad = (tile) => {
        const features = [];
        if (tile.content && tile.content.length > 0) {
            for (const feature of tile.content) {
                const ts = feature.properties.times;
                const ts_final = ts.substring(1, ts.length - 1).split(",").map(t => parseInt(t, 10)-minTimestamp);

                // slice Multi into individual features
                if (feature.geometry.type === "MultiLineString") {
                    let index = 0;
                    for (const coords of feature.geometry.coordinates) {
                        const ts_segment = ts_final.slice(index, index + coords.length)
                        features.push({
                            ...feature,
                            geometry: {type: "LineString", coordinates: coords},
                            // properties: {...feature.properties, timestamps: ts_segment}
                            properties: {timestamps: ts_segment}
                        });
                        index = coords.length;
                    }
                } else {
                    // features.push({...feature, properties: {...feature.properties, timestamps: ts_final}});
                    features.push({...feature, properties: {tripid: feature.properties.tripid, timestamps: ts_final}});
                }
            }
        }
        tile.content = features;
    };

    const trip_layer = new MVTLayer({
        id: 'trips',
        data : dataTrips,
        binary: false,
        minZoom: 0,
        maxZoom: 23,
        lineWidthMinPixels: 1,
        onTileLoad: onTileLoad,
        currentTime: time, // it has to be here, not inside the TripsLayer
        // loadOptions: {mode: 'no-cors'},
        renderSubLayers: props => {
            return new TripsLayer(props, {
                data: props.data,
                getPath: d => d.geometry.coordinates,
                getTimestamps: d => d.properties.timestamps,
                getColor: [255,255,255],
                opacity: 0.5,
                widthMinPixels: 2,
                rounded: true,
                trailLength,
            });
        }
    });

    const line_layer = new MVTLayer({
        id: 'lines',
        data : dataLines,
        binary: false,
        minZoom: 0,
        maxZoom: 23,
        lineWidthMinPixels: 1,
        onTileLoad: onTileLoad,
        currentTime: time, // it has to be here, not inside the TripsLayer
        // loadOptions: {mode: 'no-cors'},
        renderSubLayers: props => {
            // TODO: uncomment next line
            // console.log(props.data);
            // return new GeoJsonLayer(props);
            return new TripsLayer(props, {
                data: props.data,
                getPath: d => d.geometry.coordinates,
                getTimestamps: d => d.properties.timestamps,
                getColor: [255,255,255],
                opacity: 0.5,
                widthMinPixels: 2,
                rounded: true,
                trailLength,
            });
        }
    });

    function convertDateTime(input) {
        // Parse the input string into a Date object
        const date = new Date(input);

        // Inline function to pad numbers to two digits
        const pad = number => number < 10 ? '0' + number : number.toString();

        // Construct each part of the format using pad function inline
        const year = date.getFullYear();
        const month = pad(date.getMonth() + 1); // getMonth() is zero-indexed
        const day = pad(date.getDate());
        const hours = pad(date.getHours());
        const minutes = pad(date.getMinutes());
        const seconds = '00'; // Seconds are not in the input, assumed to be '00'

        // Combine parts into the desired format with the timezone offset +01
        return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`
    }


    const startAnimation = () => {
        setTime(0);
        if (!isAnimated){
            setIsAnimated(true);
        }
        if  (lineParameter !==  '' && tripParameter !== ''){
            setTripParameter('')
        }
    }
    const updateIsAnimated = () => setIsAnimated(!isAnimated);
    const updateTripParameter = () => {}
    const handleTripInputChange = (e) => setTripParameter(e.target.value);
    const updateLineParameter = () =>
        setDataLines(DATA_LINES_URL + lineParameter);
    const handleLineInputChange = (e) => setLineParameter(e.target.value);
    const handleMinTimeURLInputChange = (e) => setMinTimeURL(e.target.value);
    const handleMaxTimeURLInputChange = (e) => setMaxTimeURL(e.target.value);

    const  updateBothTimestamp = () => {

        if (minTimeURL !== '' && maxTimeURL !== '' ) {
            if (tripParameter !== '') {

                const start = convertDateTime(minTimeURL);
                const end = convertDateTime(maxTimeURL);
                setDataTrips(DATA_TRIPS_URL + tripParameter+ '&p_start=' + start + '&p_end=' + end);
            }else if (lineParameter !== '') {
                const start = convertDateTime(minTimeURL);
                const end = convertDateTime(maxTimeURL);
                setDataTrips(DATA_LINES_URL + lineParameter+ '&p_start=' + start + '&p_end=' + end);
            }
            console.log(dataTrips);
            setMinTimestamp(new Date(minTimeURL).getTime()/1000);
            setMaxTimestamp(new Date(maxTimeURL).getTime()/1000);
            console.log(minTimestamp,maxTimestamp);
            setLoopLength(maxTimestamp-minTimestamp);
            updateTripParameter()
        }
        console.log('Loop Length : ',loopLength);
    };

    const setTimeFromSlider = (newTime) => {
        setTime(newTime)
    }

    return (
        <div style={{ position: 'relative', height: '100vh' }}>
            <DeckGL
                layers={[trip_layer,line_layer]}
                initialViewState={INITIAL_VIEW_STATE}
                controller={true}
            >
                <StaticMap mapStyle={MAP_STYLE} />
            </DeckGL>
            <div style={{ position: 'absolute', top: 0, left: 0, padding: '10px' }}>
                <div>
                    <TimeSlider
                        setTimeApp={setTimeFromSlider}
                        timeFromApp={time}
                        maxfromApp={loopLength}
                        dateFromApp={new Date((minTimestamp+time)*1000)}
                    />
                </div>
                <div>
                    <input type="text" value={tripParameter} onChange={handleTripInputChange} />
                    <button onClick={updateTripParameter}>Update TripID</button>
                </div>
                <div>
                    <input type="text" value={lineParameter} onChange={handleLineInputChange} />
                    <button onClick={updateLineParameter}>Update LineID</button>
                </div>
                <div>
                    <input type="datetime-local" value={minTimeURL} onChange={handleMinTimeURLInputChange} />
                    <input type="datetime-local" value={maxTimeURL} onChange={handleMaxTimeURLInputChange} />
                    <button onClick={updateBothTimestamp}>Update Life Time</button>
                </div>
                <div>
                    <button onClick={startAnimation}>Start Animation</button>
                    <button onClick={updateIsAnimated}>
                        {isAnimated ? 'Stop Animation' : 'Replay Animation'}
                    </button>
                </div>
            </div>
        </div>
    );
}

export function renderToDOM(container) {
    render(<Router><App /></Router>, container);
}
