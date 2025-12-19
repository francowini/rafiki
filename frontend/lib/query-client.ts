import { QueryClient, DefaultOptions } from '@tanstack/react-query';

const queryConfig: DefaultOptions = {
  queries: {
    staleTime: 30 * 1000, // Data fresh for 30 seconds
    gcTime: 5 * 60 * 1000, // Keep unused data for 5 minutes
    retry: 1, // Retry failed requests once
    refetchOnWindowFocus: false, // Don't refetch on window focus
    refetchOnReconnect: true, // Refetch on reconnect
  },
  mutations: {
    retry: 1,
  },
};

export const queryClient = new QueryClient({
  defaultOptions: queryConfig,
});
