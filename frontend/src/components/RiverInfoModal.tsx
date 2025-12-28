import { RiverData } from '../types/river';
import { X, Droplet, AlertTriangle, AlertCircle, TrendingUp, Clock } from 'lucide-react';
import { format } from 'date-fns';

interface RiverInfoModalProps {
  rivers: RiverData[];
  locationName: string;
  onClose: () => void;
}

export function RiverInfoModal({ rivers, locationName, onClose }: RiverInfoModalProps) {
  if (rivers.length === 0) {
    return null;
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'safe':
        return 'bg-green-100 text-green-800 border-green-300';
      case 'caution':
        return 'bg-yellow-100 text-yellow-800 border-yellow-300';
      case 'unsafe':
        return 'bg-red-100 text-red-800 border-red-300';
      default:
        return 'bg-gray-100 text-gray-800 border-gray-300';
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'safe':
        return <Droplet className="w-5 h-5 text-green-600" />;
      case 'caution':
        return <AlertTriangle className="w-5 h-5 text-yellow-600" />;
      case 'unsafe':
        return <AlertCircle className="w-5 h-5 text-red-600" />;
      default:
        return <Droplet className="w-5 h-5 text-gray-600" />;
    }
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
      <div className="bg-white dark:bg-gray-800 rounded-xl shadow-2xl max-w-2xl w-full max-h-[90vh] overflow-hidden">
        {/* Header */}
        <div className="bg-blue-600 dark:bg-blue-700 text-white p-6 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <Droplet className="w-6 h-6" />
            <div>
              <h2 className="text-xl font-bold">River Crossing Information</h2>
              <p className="text-sm text-blue-100">{locationName}</p>
            </div>
          </div>
          <button
            onClick={onClose}
            className="hover:bg-blue-700 dark:hover:bg-blue-600 rounded-full p-2 transition-colors"
          >
            <X className="w-6 h-6" />
          </button>
        </div>

        {/* Content */}
        <div className="p-6 overflow-y-auto max-h-[calc(90vh-120px)]">
          <div className="space-y-6">
            {rivers.map((riverData, index) => (
              <div
                key={index}
                className={`border-2 rounded-lg p-4 ${getStatusColor(riverData.status)}`}
              >
                {/* River Name and Status */}
                <div className="flex items-start justify-between mb-4">
                  <div className="flex items-center gap-2">
                    {getStatusIcon(riverData.status)}
                    <div>
                      <h3 className="font-bold text-lg">{riverData.river.river_name}</h3>
                      {riverData.river.description && (
                        <p className="text-sm opacity-75">{riverData.river.description}</p>
                      )}
                    </div>
                  </div>
                  <div className="text-right">
                    <div className="font-bold text-sm uppercase">{riverData.status}</div>
                  </div>
                </div>

                {/* Status Message */}
                <div className="mb-4 p-3 bg-white bg-opacity-50 rounded-lg">
                  <p className="font-medium">{riverData.status_message}</p>
                </div>

                {/* Metrics Grid */}
                <div className="grid grid-cols-2 gap-4">
                  {/* Current Flow */}
                  <div className="bg-white bg-opacity-50 rounded-lg p-3">
                    <div className="text-xs font-medium opacity-75 mb-1">Current Flow</div>
                    <div className="flex items-baseline gap-1">
                      <span className="text-2xl font-bold">
                        {Math.round(riverData.flow_cfs)}
                      </span>
                      <span className="text-sm">CFS</span>
                    </div>
                  </div>

                  {/* Gauge Height */}
                  <div className="bg-white bg-opacity-50 rounded-lg p-3">
                    <div className="text-xs font-medium opacity-75 mb-1">Gauge Height</div>
                    <div className="flex items-baseline gap-1">
                      <span className="text-2xl font-bold">
                        {riverData.gauge_height_ft.toFixed(2)}
                      </span>
                      <span className="text-sm">ft</span>
                    </div>
                  </div>

                  {/* Safe Threshold */}
                  <div className="bg-white bg-opacity-50 rounded-lg p-3">
                    <div className="text-xs font-medium opacity-75 mb-1">Safe Threshold</div>
                    <div className="flex items-baseline gap-1">
                      <span className="text-lg font-bold">
                        {riverData.river.safe_crossing_cfs}
                      </span>
                      <span className="text-sm">CFS</span>
                    </div>
                  </div>

                  {/* Percent of Safe */}
                  <div className="bg-white bg-opacity-50 rounded-lg p-3">
                    <div className="text-xs font-medium opacity-75 mb-1 flex items-center gap-1">
                      <TrendingUp className="w-3 h-3" />
                      <span>% of Safe</span>
                    </div>
                    <div className="flex items-baseline gap-1">
                      <span className="text-2xl font-bold">
                        {Math.round(riverData.percent_of_safe)}%
                      </span>
                    </div>
                  </div>
                </div>

                {/* Timestamp */}
                <div className="mt-4 pt-3 border-t border-current border-opacity-20">
                  <div className="flex items-center gap-1 text-xs opacity-75">
                    <Clock className="w-3 h-3" />
                    <span>
                      Last updated: {format(new Date(riverData.timestamp), 'MMM d, h:mm a')}
                    </span>
                  </div>
                </div>

                {/* USGS Link */}
                <div className="mt-2">
                  <a
                    href={`https://waterdata.usgs.gov/monitoring-location/${riverData.river.gauge_id}/`}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-xs underline opacity-75 hover:opacity-100"
                  >
                    View on USGS Water Data →
                  </a>
                </div>
              </div>
            ))}
          </div>

          {/* Safety Warning */}
          <div className="mt-6 p-4 bg-gray-100 dark:bg-gray-900 rounded-lg">
            <p className="text-xs text-gray-600 dark:text-gray-400">
              <strong>Safety Note:</strong> River conditions can change rapidly. Always assess
              conditions on-site before crossing. These thresholds are estimates and may not
              account for all factors affecting crossing safety.
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}
