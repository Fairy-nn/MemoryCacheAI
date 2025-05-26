"use client";

import { useState, useEffect } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Brain, Search, Plus, AlertCircle, Database, Settings, Save, X, Trash2, MoreVertical, Loader2 } from "lucide-react";

// Types based on the API models
interface Memory {
  id?: string;
  content: string;
  score?: number;
  metadata?: {
    id?: string;
    user_id: string;
    content?: string;
    role?: string;
    source?: string;
    session_id: string;
    timestamp?: number;
    ttl: number;
  };
  timestamp: string;
}

interface MemoryStats {
  vector_db?: {
    result?: {
      vectorCount?: number;
      dimension?: number;
      indexSize?: number;
    };
  };
}

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export default function Dashboard() {
  // State management
  const [currentUser, setCurrentUser] = useState("user-001");
  const [currentSession, setCurrentSession] = useState(`session-${Date.now()}`);
  const [recentMemories, setRecentMemories] = useState<Memory[]>([]);
  const [allMemories, setAllMemories] = useState<Memory[]>([]);
  const [stats, setStats] = useState<MemoryStats | null>(null);
  
  // Form states
  const [newMemoryContent, setNewMemoryContent] = useState("");
  const [searchQuery, setSearchQuery] = useState("");
  const [queryResults, setQueryResults] = useState<Memory[]>([]);
  
  // UI states
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState("");
  const [showAllMemories, setShowAllMemories] = useState(false);
  const [showSettings, setShowSettings] = useState(false);
  const [deletingMemoryId, setDeletingMemoryId] = useState<string | null>(null);
  
  // Settings states
  const [tempUserId, setTempUserId] = useState(currentUser);
  const [minSearchScore, setMinSearchScore] = useState(0.7);
  const [searchLimit, setSearchLimit] = useState(10);
  const [memoriesLimit, setMemoriesLimit] = useState(100);

  // API functions
  const apiCall = async (endpoint: string, options: RequestInit = {}) => {
    try {
      const response = await fetch(`${API_BASE_URL}${endpoint}`, {
        headers: {
          'Content-Type': 'application/json',
          ...options.headers,
        },
        ...options,
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || `HTTP ${response.status}`);
      }

      return await response.json();
    } catch (error) {
      console.error('API call failed:', error);
      throw error;
    }
  };

  const saveMemory = async () => {
    if (!newMemoryContent.trim()) return;
    
    setIsLoading(true);
    setError("");
    
    try {
      await apiCall('/memory/save', {
        method: 'POST',
        body: JSON.stringify({
          user_id: currentUser,
          session_id: currentSession,
          content: newMemoryContent,
          role: "user"
        }),
      });
      
      setNewMemoryContent("");
      await Promise.all([fetchRecentMemories(), fetchStats(), showAllMemories && fetchAllMemories()]);
    } catch (error) {
      setError(`Failed to save memory: ${error}`);
    } finally {
      setIsLoading(false);
    }
  };

  const queryMemory = async () => {
    if (!searchQuery.trim()) return;
    
    setIsLoading(true);
    setError("");
    
    try {
      const response = await apiCall('/memory/query', {
        method: 'POST',
        body: JSON.stringify({
          user_id: currentUser,
          query: searchQuery,
          limit: searchLimit,
          min_score: minSearchScore
        }),
      });
      
      setQueryResults(response.results || []);
    } catch (error) {
      setError(`Failed to query memory: ${error}`);
    } finally {
      setIsLoading(false);
    }
  };

  const searchMemoriesByKeyword = async () => {
    if (!searchQuery.trim()) return;
    
    setIsLoading(true);
    setError("");
    
    try {
      const response = await apiCall(`/user/${currentUser}/memories/search?q=${encodeURIComponent(searchQuery)}&limit=${searchLimit}`);
      setQueryResults(response.memories || []);
    } catch (error) {
      setError(`Failed to search memories: ${error}`);
    } finally {
      setIsLoading(false);
    }
  };

  const fetchRecentMemories = async () => {
    try {
      const response = await apiCall(`/user/${currentUser}/memories/recent?limit=10`);
      setRecentMemories(response.memories || []);
    } catch (error) {
      console.error('Failed to fetch recent memories:', error);
    }
  };

  const fetchAllMemories = async () => {
    try {
      const response = await apiCall(`/user/${currentUser}/memories/recent?limit=${memoriesLimit}`);
      setAllMemories(response.memories || []);
    } catch (error) {
      console.error('Failed to fetch all memories:', error);
    }
  };

  const fetchStats = async () => {
    try {
      const response = await apiCall('/memory/stats');
      setStats(response);
    } catch (error) {
      console.error('Failed to fetch stats:', error);
    }
  };

  const deleteMemory = async (memoryId: string) => {
    console.log('Attempting to delete memory:', memoryId);
    if (!memoryId) {
      setError("Cannot delete memory: memory ID does not exist");
      return;
    }
    
    setDeletingMemoryId(memoryId);
    setError("");
    
    try {
      const response = await apiCall(`/memory/${encodeURIComponent(memoryId)}?user_id=${currentUser}`, {
        method: 'DELETE',
      });
      
      console.log('Delete memory success:', response);
      
      // Refresh memories
      await Promise.all([fetchRecentMemories(), fetchStats(), showAllMemories && fetchAllMemories()]);
      
      // Remove from query results if present
      setQueryResults(prev => prev.filter(memory => {
        const currentMemoryId = memory.id || 
                              memory.metadata?.id || 
                              (memory.metadata?.content && memory.content === memory.metadata.content ? memory.content : null);
        return currentMemoryId !== memoryId;
      }));
    } catch (error: any) {
      console.error('Delete memory failed:', error);
      setError(`Delete memory failed: ${error.message || error}`);
    } finally {
      setDeletingMemoryId(null);
    }
  };

  const cleanupUserMemories = async () => {
    if (!confirm('Are you sure you want to delete all your memories? This action cannot be undone.')) return;
    
    setIsLoading(true);
    setError("");
    
    try {
      await apiCall(`/user/${currentUser}/memories`, {
        method: 'DELETE',
      });
      
      await Promise.all([fetchRecentMemories(), fetchStats(), showAllMemories && fetchAllMemories()]);
      setQueryResults([]);
    } catch (error) {
      setError(`Failed to cleanup memories: ${error}`);
    } finally {
      setIsLoading(false);
    }
  };

  const saveSettings = () => {
    setCurrentUser(tempUserId);
    setCurrentSession(`session-${Date.now()}`);
    setShowSettings(false);
    // Reload data
    Promise.all([fetchStats(), fetchRecentMemories(), fetchAllMemories()]);
  };

  const cancelSettings = () => {
    setTempUserId(currentUser);
    setShowSettings(false);
  };

  // Load initial data
  useEffect(() => {
    fetchStats();
    fetchRecentMemories();
    fetchAllMemories();
  }, [currentUser]);

  const formatTimestamp = (timestamp: string) => {
    try {
      return new Date(timestamp).toLocaleString();
    } catch {
      return timestamp;
    }
  };

  // Memory component
  const MemoryItem = ({ memory, isSearchResult = false }: { memory: Memory, isSearchResult?: boolean }) => {
    console.log('MemoryItem', memory);  
    
    let memoryId = null;
    
    if (memory.id) {
      memoryId = memory.id;
      console.log('Using memory.id:', memoryId);
    } 
    else if (memory.metadata?.id) {
      memoryId = memory.metadata.id;
      console.log('Using memory.metadata.id:', memoryId);
    }
    else if (memory.metadata?.content && memory.content === memory.metadata.content) {
      memoryId = memory.content;
      console.log('Using content as ID:', memoryId);
    }
    else if (memory.content) {
      memoryId = memory.content;
      console.log('Using memory.content as ID:', memoryId);
    }
    
    const isDeleting = deletingMemoryId === memoryId;

    return (
      <div className="border border-gray-200 rounded-xl p-6 bg-white transition-all duration-200 hover:border-gray-300">
        <div className="flex justify-between items-start mb-3">
          <div>
            {isSearchResult && memory.score && (
              <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800">
                Score: {memory.score.toFixed(3)}
              </span>
            )}
          </div>
          <div className="flex items-center gap-3">
            <span className="text-sm text-gray-500 font-mono">
              {formatTimestamp(memory.timestamp)}
            </span>
            {memoryId && (
              <button 
                onClick={() => deleteMemory(memoryId as string)}
                disabled={isDeleting}
                className={`p-1.5 ${isDeleting ? 'cursor-not-allowed' : 'cursor-pointer'} text-gray-400 hover:text-red-500 hover:bg-red-50 rounded-lg transition-all duration-200`}
              >
                {isDeleting ? <Loader2 className="h-4 w-4 animate-spin" /> : <Trash2 className="h-4 w-4" />}
              </button>
            )}
          </div>
        </div>
        <p className="text-gray-900 leading-relaxed">{memory.content}</p>
        {memory.metadata && (
          <div className="mt-4 pt-3 border-t border-gray-100">
            <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800 font-mono">
              {memory.metadata.session_id}
            </span>
          </div>
        )}
      </div>
    );
  };

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="border-b border-gray-200 bg-white">
        <div className="max-w-7xl mx-auto px-6 py-6">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-3">
              <div className="p-2 bg-gradient-to-br from-purple-500 to-blue-600 rounded-xl">
                <Brain className="h-6 w-6 text-white" />
              </div>
              <div>
                <h1 className="text-2xl font-bold text-gray-900">Memory Cache AI</h1>
                <p className="text-sm text-gray-500 mt-0.5">Intelligent memory storage and retrieval</p>
              </div>
            </div>
            <div className="flex items-center space-x-4">
              <div className="flex items-center space-x-2 px-3 py-2 bg-gray-100 rounded-lg">
                <div className="w-2 h-2 bg-green-500 rounded-full"></div>
                <span className="text-sm font-medium text-gray-700">{currentUser}</span>
              </div>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => setShowSettings(true)}
                className="flex cursor-pointer items-center gap-2 text-gray-600 hover:text-gray-900 hover:bg-gray-100"
              >
                <Settings className="h-4 w-4" />
                Settings
              </Button>
            </div>
          </div>
        </div>
      </header>

      <div className="max-w-7xl mx-auto px-6 py-8">
        {/* Error Display */}
        {error && (
          <div className="mb-8 p-4 bg-red-50 border border-red-200 rounded-xl flex items-start gap-3">
            <AlertCircle className="h-5 w-5 text-red-500 flex-shrink-0 mt-0.5" />
            <div className="flex-1">
              <p className="text-red-800 font-medium">Error</p>
              <p className="text-red-700 text-sm mt-1">{error}</p>
            </div>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setError("")}
              className="text-red-500 cursor-pointer hover:text-red-700 hover:bg-red-100 h-6 w-6 p-0"
            >
              <X className="h-4 w-4" />
            </Button>
          </div>
        )}

        {/* Settings Panel */}
        {showSettings && (
          <div className="mb-8">
            <div className="border border-gray-200 rounded-2xl bg-white overflow-hidden">
              <div className="p-6 border-b border-gray-200 bg-gray-50">
                <div className="flex justify-between items-center">
                  <div className="flex items-center gap-3">
                    <div className="p-2 bg-gray-200 rounded-lg">
                      <Settings className="h-5 w-5 text-gray-600" />
                    </div>
                    <div>
                      <h2 className="text-xl font-semibold text-gray-900">Settings</h2>
                      <p className="text-sm text-gray-500 mt-0.5">Configure your memory preferences</p>
                    </div>
                  </div>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={cancelSettings}
                    className="text-gray-500 cursor-pointer hover:text-gray-700 hover:bg-gray-200"
                  >
                    <X className="h-4 w-4" />
                  </Button>
                </div>
              </div>
              
              <div className="p-6">
                <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
                  {/* User Settings */}
                  <div className="space-y-6">
                    <div>
                      <h3 className="text-lg font-semibold text-gray-900 mb-4">User Configuration</h3>
                      <div className="space-y-4">
                        <div>
                          <label className="block text-sm font-medium text-gray-700 mb-2">
                            User ID
                          </label>
                          <Input
                            value={tempUserId}
                            onChange={(e) => setTempUserId(e.target.value)}
                            placeholder="Enter your user ID"
                            className="w-full border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                          />
                          <p className="text-xs text-gray-500 mt-2">
                            Switch between different user profiles and memory spaces
                          </p>
                        </div>
                      </div>
                    </div>
                  </div>

                  {/* Search Settings */}
                  <div className="space-y-6">
                    <div>
                      <h3 className="text-lg font-semibold text-gray-900 mb-4">Search Configuration</h3>
                      <div className="space-y-4">
                        <div>
                          <label className="block text-sm font-medium text-gray-700 mb-2">
                            Minimum Search Score: {minSearchScore}
                          </label>
                          <input
                            type="range"
                            min="0.1"
                            max="1.0"
                            step="0.1"
                            value={minSearchScore}
                            onChange={(e) => setMinSearchScore(parseFloat(e.target.value))}
                            className="w-full h-2 bg-gray-200 rounded-lg appearance-none cursor-pointer"
                            style={{
                              background: `linear-gradient(to right, #3b82f6 0%, #3b82f6 ${(minSearchScore - 0.1) / 0.9 * 100}%, #e5e7eb ${(minSearchScore - 0.1) / 0.9 * 100}%, #e5e7eb 100%)`
                            }}
                          />
                          <p className="text-xs text-gray-500 mt-2">
                            Higher values return more relevant but fewer results
                          </p>
                        </div>

                        <div>
                          <label className="block text-sm font-medium text-gray-700 mb-2">
                            Search Results Limit
                          </label>
                          <Input
                            type="number"
                            min="1"
                            max="50"
                            value={searchLimit}
                            onChange={(e) => setSearchLimit(parseInt(e.target.value) || 10)}
                            className="w-full border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                          />
                          <p className="text-xs text-gray-500 mt-2">
                            Maximum number of search results to return
                          </p>
                        </div>

                        <div>
                          <label className="block text-sm font-medium text-gray-700 mb-2">
                            Memory Fetch Limit
                          </label>
                          <Input
                            type="number"
                            min="10"
                            max="500"
                            value={memoriesLimit}
                            onChange={(e) => setMemoriesLimit(parseInt(e.target.value) || 100)}
                            className="w-full border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                          />
                          <p className="text-xs text-gray-500 mt-2">
                            Maximum number of memories to fetch when viewing all
                          </p>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>

                <div className="flex justify-end gap-3 pt-6 border-t border-gray-200 mt-8">
                  <Button
                    variant="outline"
                    onClick={cancelSettings}
                    className="border-gray-300 cursor-pointer text-gray-700 hover:bg-gray-50"
                  >
                    Cancel
                  </Button>
                  <Button
                    onClick={saveSettings}
                    className="bg-gradient-to-r cursor-pointer from-purple-500 to-blue-600 text-white hover:from-purple-600 hover:to-blue-700 border-0"
                  >
                    <Save className="h-4 w-4 mr-2" />
                    Save Settings
                  </Button>
                </div>
              </div>
            </div>
          </div>
        )}

        <div className="space-y-8">
          <div className="space-y-8">
            {/* Add Memory Section */}
            <div className="border border-gray-200 rounded-2xl bg-white overflow-hidden">
              <div className="p-6 border-b border-gray-200 bg-gradient-to-r from-green-50 to-emerald-50">
                <div className="flex items-center gap-3">
                  <div className="p-2 bg-green-100 rounded-lg">
                    <Plus className="h-5 w-5 text-green-600" />
                  </div>
                  <div>
                    <h2 className="text-xl font-semibold text-gray-900">Create New Memory</h2>
                    <p className="text-sm text-gray-600 mt-0.5">Store information for future retrieval</p>
                  </div>
                </div>
              </div>
              <div className="p-6">
                <textarea
                  value={newMemoryContent}
                  onChange={(e) => setNewMemoryContent(e.target.value)}
                  placeholder="Enter information you want to remember..."
                  className="w-full p-4 border border-gray-300 rounded-xl bg-white text-gray-900 min-h-[120px] resize-vertical focus:ring-2 focus:ring-green-500 focus:border-transparent transition-all duration-200"
                />
                <div className="mt-4 flex justify-end">
                  <Button
                    onClick={saveMemory}
                    disabled={isLoading || !newMemoryContent.trim()}
                    className="bg-gradient-to-r cursor-pointer from-green-500 to-emerald-600 text-white hover:from-green-600 hover:to-emerald-700 border-0 px-6"
                  >
                    {isLoading ? <Loader2 className="h-4 w-4 animate-spin mr-2" /> : <Plus className="h-4 w-4 mr-2" />}
                    Save Memory
                  </Button>
                </div>
              </div>
            </div>

            {/* Search Section */}
            <div className="border border-gray-200 rounded-2xl bg-white overflow-hidden">
              <div className="p-6 border-b border-gray-200 bg-gradient-to-r from-blue-50 to-indigo-50">
                <div className="flex items-center gap-3">
                  <div className="p-2 bg-blue-100 rounded-lg">
                    <Search className="h-5 w-5 text-blue-600" />
                  </div>
                  <div>
                    <h2 className="text-xl font-semibold text-gray-900">Search Memories</h2>
                    <p className="text-sm text-gray-600 mt-0.5">Find relevant information using AI or keyword search</p>
                  </div>
                </div>
              </div>
              <div className="p-6">
                <div className="flex gap-3 mb-6">
                  <Input
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    placeholder="Search for memories..."
                    onKeyPress={(e) => e.key === 'Enter' && queryMemory()}
                    className="flex-1 border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  />
                  <Button
                    onClick={queryMemory}
                    disabled={isLoading || !searchQuery.trim()}
                    className="bg-gradient-to-r cursor-pointer from-blue-500 to-indigo-600 text-white hover:from-blue-600 hover:to-indigo-700 border-0 px-6"
                  >
                    {isLoading ? <Loader2 className="h-4 w-4 animate-spin mr-2" /> : null}
                    AI Search
                  </Button>
                  <Button
                    onClick={searchMemoriesByKeyword}
                    disabled={isLoading || !searchQuery.trim()}
                    className="border-gray-300 cursor-pointer text-gray-700 hover:bg-gray-50 px-6"
                    variant="outline"
                  >
                    {isLoading ? <Loader2 className="h-4 w-4 animate-spin mr-2" /> : null}
                    Keyword
                  </Button>
                </div>

                {/* Loading State for Search */}
                {isLoading && searchQuery && (
                  <div className="mb-6 p-6 bg-blue-50 border border-blue-200 rounded-xl">
                    <div className="flex items-center gap-3">
                      <Loader2 className="h-5 w-5 animate-spin text-blue-600" />
                      <div>
                        <p className="text-blue-800 font-medium">Searching memories...</p>
                        <p className="text-blue-600 text-sm">Looking for relevant information</p>
                      </div>
                    </div>
                  </div>
                )}

                {/* Search Results */}
                {queryResults.length > 0 && (
                  <div>
                    <div className="flex items-center gap-2 mb-4">
                      <h3 className="font-semibold text-gray-900">Search Results</h3>
                      <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
                        {queryResults.length} results
                      </span>
                    </div>
                    <div className="space-y-4 max-h-96 overflow-y-auto">
                      {queryResults.map((result, index) => (
                        <MemoryItem key={index} memory={result} isSearchResult={true} />
                      ))}
                    </div>
                  </div>
                )}
              </div>
            </div>
            {/* Stats and Memories Row */}
            <div className="grid grid-cols-1 lg:grid-cols-4 gap-8">
              {/* Stats Card */}
              <div className="lg:col-span-1">
                <div className="border border-gray-200 rounded-2xl bg-white overflow-hidden">
                  <div className="p-6 border-b border-gray-200 bg-gradient-to-r from-purple-50 to-pink-50">
                    <div className="flex items-center gap-3">
                      <div className="p-2 bg-purple-100 rounded-lg">
                        <Database className="h-5 w-5 text-purple-600" />
                      </div>
                      <div>
                        <h2 className="text-xl font-semibold text-gray-900">Memory Stats</h2>
                        <p className="text-sm text-gray-600 mt-0.5">Overview of your stored data</p>
                      </div>
                    </div>
                  </div>
                  <div className="p-6">
                    <div className="space-y-4">
                      <div className="p-4 bg-gradient-to-r from-purple-50 to-pink-50 rounded-xl">
                        <div className="text-sm font-medium text-gray-600 mb-1">Total Memories</div>
                        <div className="text-3xl font-bold text-gray-900">
                          {stats?.vector_db?.result?.vectorCount || 0}
                        </div>
                      </div>
                      <div className="p-4 bg-gradient-to-r from-blue-50 to-indigo-50 rounded-xl">
                        <div className="text-sm font-medium text-gray-600 mb-1">Recent Memories</div>
                        <div className="text-3xl font-bold text-gray-900">
                          {recentMemories.length}
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>

              {/* Memories List */}
              <div className="lg:col-span-3">
                <div className="border border-gray-200 rounded-2xl bg-white overflow-hidden">
                  <div className="p-6 border-b border-gray-200 bg-gradient-to-r from-gray-50 to-slate-50">
                    <div className="flex justify-between items-center">
                      <div className="flex items-center gap-3">
                        <div className="p-2 bg-gray-200 rounded-lg">
                          <Brain className="h-5 w-5 text-gray-600" />
                        </div>
                        <div>
                          <h2 className="text-xl font-semibold text-gray-900">
                            {showAllMemories ? 'All Memories' : 'Recent Memories'}
                          </h2>
                          <p className="text-sm text-gray-600 mt-0.5">Your stored information</p>
                        </div>
                      </div>
                      <div className="flex gap-2">
                        <Button
                          onClick={() => setShowAllMemories(!showAllMemories)}
                          variant="outline"
                          size="sm"
                          disabled={isLoading}
                          className="border-gray-300 cursor-pointer text-gray-700 hover:bg-gray-50 text-xs px-3"
                        >
                          {showAllMemories ? 'Recent' : 'View All'}
                        </Button>
                        <Button
                          onClick={cleanupUserMemories}
                          variant="outline"
                          size="sm"
                          disabled={isLoading}
                          className="border-red-200 cursor-pointer text-red-600 hover:bg-red-50 hover:border-red-300 text-xs px-3"
                        >
                          {isLoading ? <Loader2 className="h-4 w-4 animate-spin" /> : 'Clear All'}
                        </Button>
                      </div>
                    </div>
                  </div>

                  <div className="p-6">
                    {(showAllMemories ? allMemories : recentMemories).length > 0 ? (
                      <div>
                        <div className="flex items-center gap-2 mb-4">
                          <span className="text-sm text-gray-600">Showing</span>
                          <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-800">
                            {(showAllMemories ? allMemories : recentMemories).length} memories
                          </span>
                        </div>
                        <div className="space-y-4 max-h-[600px] overflow-y-auto">
                          {(showAllMemories ? allMemories : recentMemories).map((memory, index) => (
                            <MemoryItem key={index} memory={memory} />
                          ))}
                        </div>
                      </div>
                    ) : (
                      <div className="text-center py-12">
                        <div className="p-4 bg-gray-100 rounded-2xl inline-block mb-4">
                          <Brain className="h-12 w-12 text-gray-400" />
                        </div>
                        <h3 className="text-lg font-semibold text-gray-700 mb-2">
                          No memories found
                        </h3>
                        <p className="text-gray-500 text-sm">
                          Start adding memories to see them here.
                        </p>
                      </div>
                    )}
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}