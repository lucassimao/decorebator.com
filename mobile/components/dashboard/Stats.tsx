export interface DashboardStats {
  totalWords: number;
  wordlistCount: number;
  wordsLearned: number;
}

import { useQuery } from "@tanstack/react-query";
import React from "react";
import {
  StyleSheet,
  Text,
  View
} from "react-native";


const fetchDashboardStats = async (): Promise<DashboardStats> => {
  // Simulate API call
  await new Promise((resolve) => setTimeout(resolve, 500));

  // Replace with actual API call
  return {
    totalWords: 165,
    wordlistCount: 2,
    wordsLearned: 89,
  };
};
type DashboardStatsProps={}
const DashboardStats: React.FC<DashboardStatsProps> = () => {

  // Fetch stats
  const { data: stats ,isLoading} = useQuery({
    queryKey: ["dashboardStats"],
    queryFn: fetchDashboardStats,
  });



  return (
    <View style={styles.statsContainer}>
           <View style={styles.statItem}>
             <Text style={styles.statLabel}>Total Words</Text>
             <Text style={styles.statValue}>{stats?.totalWords || 0}</Text>
           </View>
           <View style={styles.statItem}>
             <Text style={styles.statLabel}>Wordlists</Text>
             <Text style={styles.statValue}>{stats?.wordlistCount || 0}</Text>
           </View>
           <View style={styles.statItem}>
             <Text style={styles.statLabel}>Words Learned</Text>
             <Text style={styles.statValue}>{stats?.wordsLearned || 0}</Text>
           </View>
         </View>
  );
};

export default DashboardStats;

const styles = StyleSheet.create({

  statsContainer: {
    flexDirection: "row",
    justifyContent: "space-between",
    backgroundColor: "rgba(255, 255, 255, 0.7)",
    borderRadius: 16,
    paddingVertical: 20,
    paddingHorizontal: 10,
    marginHorizontal: 20,
    marginBottom: 24,
    shadowColor: "#000",
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.05,
    shadowRadius: 8,
    elevation: 2,
  },
  statItem: {
    alignItems: "center",
    flex: 1,
  },
  statLabel: {
    fontSize: 14,
    color: "#636E72",
    marginBottom: 8,
  },
  statValue: {
    fontSize: 36,
    fontWeight: "700",
    color: "#2D3436",
  },
});
