import { WeatherForecast } from '../types/weather';
import {
  getWeatherCondition,
  getConditionColor,
  getWindDirection,
  getWeatherIconUrl,
  calculate48HourRain,
  getSnowProbability
} from '../utils/weatherConditions';
import { format } from 'date-fns';
import { Cloud, Droplet, Wind, Snowflake, ChevronDown } from 'lucide-react';

interface WeatherCardProps {
  forecast: WeatherForecast;
  isExpanded: boolean;
  onToggleExpand: (expanded: boolean) => void;
}

export function WeatherCard({ forecast, isExpanded, onToggleExpand }: WeatherCardProps) {
  const { location, current, hourly, historical } = forecast;
  const condition = getWeatherCondition(current);
  const conditionColor = getConditionColor(condition.level);

  // Safely handle potentially null/undefined arrays
  const safeHistorical = historical || [];
  const safeHourly = hourly || [];

  // Calculate past 48-hour rain (total)
  const allData = [...safeHistorical, ...safeHourly].sort((a, b) =>
    new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime()
  );
  const now = new Date();
  const past48h = allData.filter(d => {
    const time = new Date(d.timestamp).getTime();
    const fortyEightHoursAgo = now.getTime() - 48 * 60 * 60 * 1000;
    return time >= fortyEightHoursAgo && time <= now.getTime();
  });
  const rainLast48h = past48h.reduce((sum, d) => sum + d.precipitation, 0);

  // Calculate next 48-hour rain forecast (average per day)
  const next48h = safeHourly.filter(d => {
    const time = new Date(d.timestamp).getTime();
    const fortyEightHoursFromNow = now.getTime() + 48 * 60 * 60 * 1000;
    return time > now.getTime() && time <= fortyEightHoursFromNow;
  });
  const totalRainNext48h = next48h.reduce((sum, d) => sum + d.precipitation, 0);
  const rainNext48h = totalRainNext48h / 2; // Average per day

  // Check for snow on ground (use recent historical data)
  const recentData = [...safeHistorical].reverse().slice(0, 8); // Last 24 hours
  const snowInfo = getSnowProbability(recentData);

  // Determine precipitation type for last 48h
  const hasSnowLast48h = past48h.some(d => d.temperature <= 32 && d.precipitation > 0);
  const hasRainLast48h = past48h.some(d => d.temperature > 32 && d.precipitation > 0);

  // Determine precipitation type for next 48h
  const hasSnowNext48h = next48h.some(d => d.temperature <= 32 && d.precipitation > 0);
  const hasRainNext48h = next48h.some(d => d.temperature > 32 && d.precipitation > 0);

  // Determine if rain is bad (for last 48h total, for next 48h avg per day)
  const rainLast48hBad = rainLast48h > 2; // Total over 2 days
  const rainNext48hBad = rainNext48h > 1; // Average per day

  return (
    <div className="bg-white rounded-xl shadow-md hover:shadow-lg transition-shadow duration-200 border border-gray-200">
      {/* Main Card Content */}
      <div className="p-6">
        {/* Header */}
        <div className="flex items-center justify-between mb-4">
          <div>
            <h2 className="text-2xl font-bold text-gray-900">{location.name}</h2>
            <p className="text-sm text-gray-500 mt-1">
              {format(new Date(current.timestamp), 'MMM d, h:mm a')}
            </p>
          </div>
          <div className={`w-4 h-4 rounded-full ${conditionColor} shadow-sm`} title={condition.level} />
        </div>

        {/* Current Weather */}
        <div className="flex items-center gap-4 mb-6">
          <img
            src={getWeatherIconUrl(current.icon)}
            alt={current.description}
            className="w-20 h-20"
          />
          <div>
            <div className="text-4xl font-bold text-gray-900">
              {Math.round(current.temperature)}°F
            </div>
            <div className="text-sm text-gray-600 capitalize mt-1">
              {current.description}
            </div>
          </div>
        </div>

        {/* Rain Alert */}
        {(rainLast48hBad || rainNext48hBad) && (
          <div className="mb-4 bg-red-50 border border-red-200 rounded-lg p-3">
            <div className="flex items-center gap-2">
              <Droplet className="w-4 h-4 text-red-600" />
              <span className="text-sm font-semibold text-red-900">
                {rainLast48hBad && `${rainLast48h.toFixed(2)}" last 48h`}
                {rainLast48hBad && rainNext48hBad && ' • '}
                {rainNext48hBad && `${rainNext48h.toFixed(2)}" next 48h`}
              </span>
            </div>
          </div>
        )}

        {/* Metrics Grid - 3 columns */}
        <div className="grid grid-cols-3 gap-3 mb-4">
          {/* Last 48h Rain/Snow */}
          <div className="flex flex-col items-center text-center">
            {hasSnowLast48h && hasRainLast48h ? (
              <div className="relative w-5 h-5 mb-1">
                <Snowflake className={`w-3 h-3 absolute top-0 left-0 ${rainLast48hBad ? 'text-red-500' : 'text-blue-400'}`} />
                <Droplet className={`w-3 h-3 absolute bottom-0 right-0 ${rainLast48hBad ? 'text-red-500' : 'text-blue-500'}`} />
              </div>
            ) : hasSnowLast48h ? (
              <Snowflake className={`w-5 h-5 mb-1 ${rainLast48hBad ? 'text-red-500' : 'text-blue-400'}`} />
            ) : (
              <Droplet className={`w-5 h-5 mb-1 ${rainLast48hBad ? 'text-red-500' : 'text-blue-500'}`} />
            )}
            <div className="text-xs text-gray-500 mb-1">Last 48h</div>
            <div className={`text-sm font-semibold ${rainLast48hBad ? 'text-red-900' : 'text-gray-900'}`}>
              {rainLast48h.toFixed(2)}"
            </div>
          </div>

          {/* Next 48h Rain/Snow */}
          <div className="flex flex-col items-center text-center">
            {hasSnowNext48h && hasRainNext48h ? (
              <div className="relative w-5 h-5 mb-1">
                <Snowflake className={`w-3 h-3 absolute top-0 left-0 ${rainNext48hBad ? 'text-red-500' : 'text-blue-400'}`} />
                <Droplet className={`w-3 h-3 absolute bottom-0 right-0 ${rainNext48hBad ? 'text-red-500' : 'text-blue-500'}`} />
              </div>
            ) : hasSnowNext48h ? (
              <Snowflake className={`w-5 h-5 mb-1 ${rainNext48hBad ? 'text-red-500' : 'text-blue-400'}`} />
            ) : (
              <Droplet className={`w-5 h-5 mb-1 ${rainNext48hBad ? 'text-red-500' : 'text-blue-400'}`} />
            )}
            <div className="text-xs text-gray-500 mb-1">Next 48h</div>
            <div className={`text-sm font-semibold ${rainNext48hBad ? 'text-red-900' : 'text-gray-900'}`}>
              {rainNext48h.toFixed(2)}"
            </div>
          </div>

          {/* Snow on Ground */}
          <div className="flex flex-col items-center text-center">
            <Snowflake className="w-5 h-5 mb-1 text-blue-400" />
            <div className="text-xs text-gray-500 mb-1">Snow</div>
            <div className="text-sm font-semibold text-gray-900">
              {snowInfo.probability}
            </div>
          </div>

          {/* Wind */}
          <div className="flex flex-col items-center text-center">
            <Wind className="w-5 h-5 mb-1 text-gray-600" />
            <div className="text-xs text-gray-500 mb-1">Wind</div>
            <div className="text-sm font-semibold text-gray-900">
              {Math.round(current.wind_speed)} {getWindDirection(current.wind_direction)}
            </div>
          </div>

          {/* Humidity */}
          <div className="flex flex-col items-center text-center">
            <Droplet className="w-5 h-5 mb-1 text-cyan-500" />
            <div className="text-xs text-gray-500 mb-1">Humid</div>
            <div className="text-sm font-semibold text-gray-900">
              {current.humidity}%
            </div>
          </div>

          {/* Cloud Cover */}
          <div className="flex flex-col items-center text-center">
            <Cloud className="w-5 h-5 mb-1 text-gray-500" />
            <div className="text-xs text-gray-500 mb-1">Clouds</div>
            <div className="text-sm font-semibold text-gray-900">
              {current.cloud_cover}%
            </div>
          </div>
        </div>

        {/* Condition Reasons */}
        <div className="border-t border-gray-200 pt-4">
          <div className="text-xs text-gray-600">
            {condition.reasons.join(' • ')}
          </div>
        </div>
      </div>

      {/* Expandable Forecast Section */}
      <button
        onClick={() => onToggleExpand(!isExpanded)}
        className="w-full px-6 py-3 border-t border-gray-200 flex items-center justify-center gap-2 text-sm font-medium text-gray-700 hover:bg-gray-50 transition-colors"
      >
        <>
          <ChevronDown className="w-4 h-4" />
          Show 6-Day Forecast
        </>
</button>
    </div>
  );
}
