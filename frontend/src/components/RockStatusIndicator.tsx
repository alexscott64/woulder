import { AlertTriangle, Droplet, Clock, CheckCircle, Stone } from 'lucide-react';
import { RockDryingStatus } from '../types/weather';

interface RockStatusIndicatorProps {
  status: RockDryingStatus;
  onClick?: () => void;
  compact?: boolean; // For small icon-only display
}

export function RockStatusIndicator({ status, onClick, compact = false }: RockStatusIndicatorProps) {
  // Determine icon color and background based on status
  const getStatusStyles = () => {
    switch (status.status) {
      case 'critical':
        return {
          iconColor: 'text-red-600 dark:text-red-400',
          bgColor: 'bg-red-50 dark:bg-red-900/20',
          borderColor: 'border-red-200 dark:border-red-800',
          hoverBg: 'hover:bg-red-100 dark:hover:bg-red-900/30',
          dotColor: 'bg-red-500',
        };
      case 'poor':
        return {
          iconColor: 'text-orange-600 dark:text-orange-400',
          bgColor: 'bg-orange-50 dark:bg-orange-900/20',
          borderColor: 'border-orange-200 dark:border-orange-800',
          hoverBg: 'hover:bg-orange-100 dark:hover:bg-orange-900/30',
          dotColor: 'bg-orange-500',
        };
      case 'fair':
        return {
          iconColor: 'text-yellow-600 dark:text-yellow-400',
          bgColor: 'bg-yellow-50 dark:bg-yellow-900/20',
          borderColor: 'border-yellow-200 dark:border-yellow-800',
          hoverBg: 'hover:bg-yellow-100 dark:hover:bg-yellow-900/30',
          dotColor: 'bg-yellow-500',
        };
      case 'good':
        return {
          iconColor: 'text-green-600 dark:text-green-400',
          bgColor: 'bg-green-50 dark:bg-green-900/20',
          borderColor: 'border-green-200 dark:border-green-800',
          hoverBg: 'hover:bg-green-100 dark:hover:bg-green-900/30',
          dotColor: 'bg-green-500',
        };
    }
  };

  const styles = getStatusStyles();

  // Render overlay badge for compact mode
  const getOverlayBadge = () => {
    switch (status.status) {
      case 'critical':
        return (
          <div className="absolute -top-1 -right-1 bg-red-500 rounded-full p-0.5 border-2 border-white dark:border-gray-800">
            <AlertTriangle className="w-2.5 h-2.5 text-white" strokeWidth={3} />
          </div>
        );
      case 'poor':
        return (
          <div className="absolute -top-1 -right-1 bg-orange-500 rounded-full p-0.5 border-2 border-white dark:border-gray-800">
            <Droplet className="w-2.5 h-2.5 text-white" fill="currentColor" />
          </div>
        );
      case 'fair':
        return (
          <div className="absolute -top-1 -right-1 bg-yellow-500 rounded-full p-0.5 border-2 border-white dark:border-gray-800">
            <Clock className="w-2.5 h-2.5 text-white" strokeWidth={2.5} />
          </div>
        );
      case 'good':
        return (
          <div className="absolute -top-1 -right-1 bg-green-500 rounded-full p-0.5 border-2 border-white dark:border-gray-800">
            <CheckCircle className="w-2.5 h-2.5 text-white" strokeWidth={2.5} />
          </div>
        );
    }
  };

  if (compact) {
    return (
      <button
        onClick={onClick}
        className={`relative p-1.5 rounded-full transition-colors ${styles.hoverBg} active:scale-95`}
        title={status.message}
      >
        <Stone className={`w-5 h-5 ${styles.iconColor}`} strokeWidth={2} />
        {getOverlayBadge()}
      </button>
    );
  }

  return (
    <button
      onClick={onClick}
      className={`flex items-center gap-3 px-3 py-2.5 rounded-lg border transition-all ${styles.bgColor} ${styles.borderColor} ${styles.hoverBg} hover:shadow-sm active:scale-98`}
    >
      <div className="relative flex-shrink-0">
        <Stone className={`w-6 h-6 ${styles.iconColor}`} strokeWidth={2} />
        <div className={`absolute -bottom-0.5 -right-0.5 w-3 h-3 rounded-full border-2 border-white dark:border-gray-800 ${styles.dotColor}`} />
      </div>
      <div className="text-left flex-1 min-w-0">
        <div className="text-sm font-semibold text-gray-900 dark:text-white truncate">
          {status.primary_group_name}
        </div>
        <div className={`text-xs font-medium ${
          status.status === 'critical' ? 'text-red-700 dark:text-red-300' :
          status.status === 'poor' ? 'text-orange-700 dark:text-orange-300' :
          status.status === 'fair' ? 'text-yellow-700 dark:text-yellow-300' :
          'text-green-700 dark:text-green-300'
        }`}>
          {status.is_wet && status.hours_until_dry > 0 ? (
            `${Math.ceil(status.hours_until_dry)}h until dry`
          ) : status.status === 'critical' ? (
            'DO NOT CLIMB'
          ) : status.status === 'good' ? (
            'Dry & Safe'
          ) : (
            status.message
          )}
        </div>
      </div>
    </button>
  );
}
