import React, { useState } from 'react';
import { Terminal } from 'lucide-react';

const GitHubProjectsTUI = () => {
  const [currentView, setCurrentView] = useState('board');
  const [selectedCard, setSelectedCard] = useState({ col: 0, row: 0 });

  // サンプルデータ
  const boardData = {
    columns: [
      { name: 'Backlog', items: ['#123 ユーザー認証機能', '#124 API統合', '#125 ドキュメント更新'] },
      { name: 'In Progress', items: ['#126 TUIデザイン実装', '#127 キーバインド設定'] },
      { name: 'Review', items: ['#128 テストコード追加'] },
      { name: 'Done', items: ['#129 初期セットアップ', '#130 README作成'] }
    ]
  };

  const tableData = [
    { id: '#123', title: 'ユーザー認証機能', status: 'Backlog', assignee: '@tanaka', priority: 'High', updated: '2024-12-01' },
    { id: '#124', title: 'API統合', status: 'Backlog', assignee: '@sato', priority: 'Medium', updated: '2024-12-02' },
    { id: '#126', title: 'TUIデザイン実装', status: 'In Progress', assignee: '@tanaka', priority: 'High', updated: '2024-12-05' },
    { id: '#127', title: 'キーバインド設定', status: 'In Progress', assignee: '@yamada', priority: 'Low', updated: '2024-12-06' },
    { id: '#128', title: 'テストコード追加', status: 'Review', assignee: '@sato', priority: 'Medium', updated: '2024-12-04' }
  ];

  const roadmapData = [
    { task: 'ユーザー認証機能 #123', sprint: 'Sprint 1', progress: '█████░░░░░', percent: '50%' },
    { task: 'TUIデザイン実装 #126', sprint: 'Sprint 1', progress: '███████░░░', percent: '70%' },
    { task: 'API統合 #124', sprint: 'Sprint 2', progress: '░░░░░░░░░░', percent: '0%' },
    { task: 'テストコード追加 #128', sprint: 'Sprint 2', progress: '████░░░░░░', percent: '40%' }
  ];

  const renderHeader = () => (
    <div className="bg-gray-900 text-green-400 px-4 py-2 font-mono text-sm border-b border-gray-700">
      <div className="flex justify-between">
        <div className="flex items-center gap-4">
          <span className="text-green-500 font-bold">█ GitHub Projects TUI</span>
          <span className="text-gray-500">|</span>
          <span className="text-blue-400">Project: Web App v2.0</span>
        </div>
        <div className="flex gap-4">
          <span className={currentView === 'board' ? 'text-yellow-400 font-bold' : 'text-gray-500'}>
            [1:Board]
          </span>
          <span className={currentView === 'table' ? 'text-yellow-400 font-bold' : 'text-gray-500'}>
            [2:Table]
          </span>
          <span className={currentView === 'roadmap' ? 'text-yellow-400 font-bold' : 'text-gray-500'}>
            [3:Roadmap]
          </span>
        </div>
      </div>
    </div>
  );

  const renderFooter = () => (
    <div className="bg-gray-900 text-gray-400 px-4 py-2 font-mono text-xs border-t border-gray-700">
      <div className="flex justify-between">
        <span className="text-green-500">NORMAL MODE</span>
        <span className="text-gray-500">
          j/k:移動 h/l:列移動 i:編集 /:フィルタ a:アサイン 1-3:ビュー切替 q:終了
        </span>
      </div>
    </div>
  );

  const renderBoardView = () => (
    <div className="flex gap-4 h-full overflow-x-auto">
      {boardData.columns.map((col, colIdx) => (
        <div key={colIdx} className="flex-shrink-0 w-64">
          <div className="bg-gray-800 text-blue-300 px-3 py-2 font-mono text-sm font-bold border border-gray-700 rounded-t">
            {col.name} ({col.items.length})
          </div>
          <div className="space-y-2 mt-2">
            {col.items.map((item, rowIdx) => (
              <div
                key={rowIdx}
                className={`bg-gray-800 border ${
                  selectedCard.col === colIdx && selectedCard.row === rowIdx
                    ? 'border-yellow-400 bg-gray-700'
                    : 'border-gray-700'
                } rounded px-3 py-2 font-mono text-sm text-gray-300 cursor-pointer hover:border-gray-500`}
                onClick={() => setSelectedCard({ col: colIdx, row: rowIdx })}
              >
                <div className="text-green-400 text-xs mb-1">{item.split(' ')[0]}</div>
                <div className="text-gray-200">{item.split(' ').slice(1).join(' ')}</div>
                <div className="flex gap-2 mt-2 text-xs">
                  <span className="text-purple-400">@user</span>
                  <span className="text-cyan-400">[bug]</span>
                </div>
              </div>
            ))}
          </div>
        </div>
      ))}
    </div>
  );

  const renderTableView = () => (
    <div className="font-mono text-sm overflow-auto">
      <table className="w-full border-collapse">
        <thead>
          <tr className="bg-gray-800 text-blue-300">
            <th className="border border-gray-700 px-3 py-2 text-left">ID</th>
            <th className="border border-gray-700 px-3 py-2 text-left">Title</th>
            <th className="border border-gray-700 px-3 py-2 text-left">Status</th>
            <th className="border border-gray-700 px-3 py-2 text-left">Assignee</th>
            <th className="border border-gray-700 px-3 py-2 text-left">Priority</th>
            <th className="border border-gray-700 px-3 py-2 text-left">Updated</th>
          </tr>
        </thead>
        <tbody>
          {tableData.map((row, idx) => (
            <tr
              key={idx}
              className={`${
                idx === selectedCard.row ? 'bg-gray-700 text-yellow-400' : 'bg-gray-900 text-gray-300'
              } hover:bg-gray-800 cursor-pointer`}
              onClick={() => setSelectedCard({ ...selectedCard, row: idx })}
            >
              <td className="border border-gray-700 px-3 py-2 text-green-400">{row.id}</td>
              <td className="border border-gray-700 px-3 py-2">{row.title}</td>
              <td className="border border-gray-700 px-3 py-2 text-cyan-400">{row.status}</td>
              <td className="border border-gray-700 px-3 py-2 text-purple-400">{row.assignee}</td>
              <td className="border border-gray-700 px-3 py-2">
                <span className={row.priority === 'High' ? 'text-red-400' : row.priority === 'Medium' ? 'text-yellow-400' : 'text-green-400'}>
                  {row.priority}
                </span>
              </td>
              <td className="border border-gray-700 px-3 py-2 text-gray-500">{row.updated}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );

  const renderRoadmapView = () => (
    <div className="font-mono text-sm space-y-3">
      <div className="bg-gray-800 text-blue-300 px-4 py-2 border border-gray-700 rounded">
        <span className="font-bold">Timeline: Sprint 1 - Sprint 2</span>
        <span className="text-gray-500 ml-4">(2024-12-01 → 2024-12-31)</span>
      </div>
      
      <div className="space-y-2">
        {roadmapData.map((item, idx) => (
          <div
            key={idx}
            className={`bg-gray-800 border ${
              idx === selectedCard.row ? 'border-yellow-400 bg-gray-700' : 'border-gray-700'
            } rounded px-4 py-3 cursor-pointer hover:border-gray-500`}
            onClick={() => setSelectedCard({ ...selectedCard, row: idx })}
          >
            <div className="flex justify-between items-center mb-2">
              <span className="text-gray-200">{item.task}</span>
              <span className="text-cyan-400 text-xs">{item.sprint}</span>
            </div>
            <div className="flex items-center gap-3">
              <span className="text-green-400 text-lg">{item.progress}</span>
              <span className="text-yellow-400">{item.percent}</span>
            </div>
          </div>
        ))}
      </div>

      <div className="bg-gray-800 border border-gray-700 rounded px-4 py-3 mt-4">
        <div className="text-gray-400 text-xs mb-2">Sprint Progress Overview:</div>
        <div className="space-y-1">
          <div className="flex items-center gap-3">
            <span className="text-cyan-400 w-24">Sprint 1:</span>
            <span className="text-green-400">████████░░</span>
            <span className="text-yellow-400">60%</span>
          </div>
          <div className="flex items-center gap-3">
            <span className="text-cyan-400 w-24">Sprint 2:</span>
            <span className="text-green-400">██░░░░░░░░</span>
            <span className="text-yellow-400">20%</span>
          </div>
        </div>
      </div>
    </div>
  );

  return (
    <div className="min-h-screen bg-black text-green-400 flex flex-col">
      {renderHeader()}
      
      <div className="flex-1 overflow-hidden">
        <div className="h-full p-4 overflow-auto">
          {currentView === 'board' && renderBoardView()}
          {currentView === 'table' && renderTableView()}
          {currentView === 'roadmap' && renderRoadmapView()}
        </div>
      </div>

      {renderFooter()}

      {/* ビュー切り替えボタン（デモ用） */}
      <div className="fixed top-20 right-4 bg-gray-800 border border-gray-700 rounded p-3 space-y-2">
        <div className="text-gray-400 text-xs mb-2 font-mono">デモ操作:</div>
        <button
          onClick={() => setCurrentView('board')}
          className="w-full px-3 py-1 bg-gray-700 hover:bg-gray-600 text-green-400 rounded font-mono text-sm"
        >
          [1] Board
        </button>
        <button
          onClick={() => setCurrentView('table')}
          className="w-full px-3 py-1 bg-gray-700 hover:bg-gray-600 text-green-400 rounded font-mono text-sm"
        >
          [2] Table
        </button>
        <button
          onClick={() => setCurrentView('roadmap')}
          className="w-full px-3 py-1 bg-gray-700 hover:bg-gray-600 text-green-400 rounded font-mono text-sm"
        >
          [3] Roadmap
        </button>
      </div>
    </div>
  );
};

export default GitHubProjectsTUI;